package telegram

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/db/tokensdb"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/db/usersdb"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/ocr"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/openai"
	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/splitwise"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"go.uber.org/zap"

	"github.com/looplab/fsm"
)

type Config struct {
	TelegramBotToken string `env:"TELEGRAM_BOT_TOKEN"`
	WebAppUrl        string `env:"WEB_APP_URL"`
}

type BotDeps struct {
	UsersDb   usersdb.Client
	TokensDb  tokensdb.Client
	Ocr       ocr.Client
	Splitwise splitwise.Client
	OpenAI    openai.Client
}

type client struct {
	deps       *BotDeps
	bot        *gotgbot.Bot
	dispatcher *ext.Dispatcher
	log        *zap.SugaredLogger
	fsms       map[int64]*fsm.FSM
	webAppUrl  string
}

const (
	eventAuthorize          = "EventAuthorize"
	eventOAuthCode          = "EventOAuthCode"
	eventCancel             = "EventCancel"
	eventSetGroup           = "EventSetGroup"
	eventSetCur             = "EventSetCur"
	eventUpload             = "EventUpload"
	eventAlterTranscription = "EventAlterTranscription"
	eventSplit              = "EventSplit"
	eventAlterSplit         = "EventAlterSplit"
	eventUnauthorize        = "EventUnauthorize"
	eventBadProfile         = "EventBadProfile"
	eventWaitForGPT         = "EventWaitForGPT"
	eventResponseFromGPT    = "EventResponseFromGPT"
	eventApprove            = "EventApprove"

	stateUnauthorized            = "unauthorized"
	stateIncompleteProfile       = "incomplete_profile"
	stateAwaitingOAuthCode       = "awaiting_oauth_code"
	stateReady                   = "ready"
	stateUploaded                = "uploaded"
	stateUploadedWaiting         = "uploaded_waiting"
	stateClarifyingTranscription = "clarifying_transcription"
	stateSplitting               = "splitting"
	stateSplittingWaiting        = "splitting_waiting"
)

func NewClient(cfg Config, deps *BotDeps) (Client, error) {
	log := zap.S()
	log.Debugln("Creating bot client")

	bot, err := gotgbot.NewBot(cfg.TelegramBotToken, nil)
	if err != nil {
		return nil, err
	}

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		// If an error is returned by a handler, log it and continue going.
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Errorw("An error occurred while handling update", "update", ctx.Update, zap.Error(err))
			b.SendMessage(ctx.EffectiveChat.Id, "Что-то пошло не так :(", &gotgbot.SendMessageOpts{})
			return ext.DispatcherActionEndGroups
		},
		Panic: func(b *gotgbot.Bot, ctx *ext.Context, r interface{}) {
			log.Errorw("A panic occurred while handling update", "update", ctx.Update, zap.Any("panic", r))
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})
	client := &client{
		deps:       deps,
		bot:        bot,
		dispatcher: dispatcher,
		log:        log,
		fsms:       make(map[int64]*fsm.FSM),
		webAppUrl:  cfg.WebAppUrl,
	}

	dispatcher.AddHandlerToGroup(handlers.NewMessage(filterUserUpdates, client.preprocess), -1)
	dispatcher.AddHandlerToGroup(handlers.NewCallback(func(cq *gotgbot.CallbackQuery) bool { return true }, client.preprocess), -1)

	dispatcher.AddHandler(handlers.NewCommand("start", client.start))
	dispatcher.AddHandler(handlers.NewCommand("help", client.help))
	dispatcher.AddHandler(handlers.NewCommand("authorize", client.authorize))
	dispatcher.AddHandler(handlers.NewCommand("set_group", client.setGroup))
	dispatcher.AddHandler(handlers.NewCommand("set_currency", client.setCurrency))

	dispatcher.AddHandler(handlers.NewCallback(func(cq *gotgbot.CallbackQuery) bool { return true }, client.callbackQuery))

	dispatcher.AddHandler(handlers.NewMessage(newMessageFilter, client.newMessage))

	dispatcher.AddHandlerToGroup(handlers.NewMessage(filterUserUpdates, client.postprocess), 1000)

	return client, nil
}

func filterUserUpdates(message *gotgbot.Message) bool {
	zap.S().Debugw("filtering update", "tg_message", message)
	return message != nil && message.From != nil && !message.From.IsBot
}

func (c *client) preprocess(b *gotgbot.Bot, ctx *ext.Context) error {
	c.log.Debug("Preprocessing user update")
	user, err := c.deps.UsersDb.GetUser(context.Background(), ctx.EffectiveUser.Id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	var userFsm *fsm.FSM = nil

	ctx.Data["user"] = user
	ctx.Data["user_dirty"] = false

	if user == nil {
		return nil
	}

	userFsm, ok := c.fsms[user.TelegramId]
	if !ok {
		userFsm = fsm.NewFSM(
			user.State,
			fsm.Events{
				{
					Name: eventAuthorize,
					Src: []string{
						stateUnauthorized,
						stateReady,
					},
					Dst: stateAwaitingOAuthCode,
				},
				{
					Name: eventCancel,
					Src: []string{
						stateAwaitingOAuthCode,
					},
					Dst: stateUnauthorized,
				},
				{
					Name: eventCancel,
					Src: []string{
						stateUploaded,
						stateSplitting,
					},
					Dst: stateReady,
				},
				{
					Name: eventSetGroup,
					Src: []string{
						stateReady,
					},
					Dst: stateReady,
				},
				{
					Name: eventSetGroup,
					Src: []string{
						stateReady,
					},
					Dst: stateReady,
				},
				{
					Name: eventUpload,
					Src: []string{
						stateReady,
					},
					Dst: stateUploaded,
				},
			},
			fsm.Callbacks{},
		)
		c.fsms[user.TelegramId] = userFsm
	}

	ctx.Data["user_fsm"] = userFsm

	if !user.Authorized {
		return nil
	}

	splitwiseInstance, ok := c.deps.Splitwise.GetInstance(user.TelegramId)
	if !ok {
		token, err := c.deps.TokensDb.GetToken(context.Background(), user.TelegramId)
		if err != nil {
			err = fmt.Errorf("failed to get token: %w", err)
		} else {
			if token == nil {
				err = fmt.Errorf("token not found")
			} else {
				splitwiseInstance, err = c.deps.Splitwise.AddInstanceFromOAuthToken(context.Background(), user.TelegramId, token.Token)
				if err != nil {
					err = fmt.Errorf("failed to add instance from oauth token: %w", err)
				}
			}
		}
		if err != nil {
			user.Authorized = false
			user.SplitwiseGroupId = -1
			user.State = stateUnauthorized
			b.SendMessage(ctx.EffectiveChat.Id, "Авторизуйтесь", &gotgbot.SendMessageOpts{})
			c.log.Errorw("User is authorized but splitwise instance couldn't be created: %w", zap.Error(err))
			return nil
		}
		if user.SplitwiseGroupId != -1 {
			_, err = splitwiseInstance.GetGroup(context.Background(), int(user.SplitwiseGroupId))
			if err != nil {
				user.SplitwiseGroupId = -1
				user.State = stateReady
				_, err = c.deps.UsersDb.UpdateUser(context.Background(), user)
				if err != nil {
					return fmt.Errorf("failed to update user: %w", err)
				}
				b.SendMessage(
					ctx.EffectiveChat.Id,
					"Произошла ошибка при получении информации о выбранной группе, повторите выбор группы",
					&gotgbot.SendMessageOpts{},
				)
			}
		}
	}
	ctx.Data["splitwise_instance"] = splitwiseInstance

	return nil
}

func (c *client) postprocess(b *gotgbot.Bot, ctx *ext.Context) error {
	c.log.Debug("Postprocessing user update")

	user := ctx.Data["user"].(*usersdb.User)
	userDirty := ctx.Data["user_dirty"].(bool)

	if !userDirty {
		return nil
	}

	c.log.Debug("User state is dirty")
	_, err := c.deps.UsersDb.UpdateUser(context.Background(), user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (c *client) makeUserProfileString(user *usersdb.User) string {
	var authorizedStr, groupString, currencyStr string

	if user.Authorized {
		authorizedStr = "Да"
	} else {
		authorizedStr = "Нет"
	}

	if user.SplitwiseGroupId == -1 {
		groupString = "Не выбрана"
	} else {
		splitwiseInstance, _ := c.deps.Splitwise.GetInstance(user.TelegramId)
		group, _ := splitwiseInstance.GetGroup(context.Background(), int(user.SplitwiseGroupId))
		groupString = group.Name
	}

	if user.Currency == "" {
		currencyStr = "Не выбрана"
	} else {
		currencyStr = user.Currency
	}

	return fmt.Sprintf(
		"Авторизован: %s\nГруппа: %s\nВалюта: %s\n",
		authorizedStr, groupString, currencyStr,
	)
}

func (c *client) HandleUpdate(ctx context.Context, update *gotgbot.Update) error {
	c.log.Debugw("Handling update", "update", update)
	return c.dispatcher.ProcessUpdate(c.bot, update, map[string]interface{}{"ctx": ctx})
}

func (c *client) Auth(ctx context.Context, code string, state string) error {
	c.log.Debugw("Handling splitwise callback", "state", state)

	telegramId, _, err := parseState(state)
	if err != nil {
		return fmt.Errorf("failed to parse state: %w", err)
	}

	user, err := c.deps.UsersDb.GetUser(ctx, telegramId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.SplitwiseOAuthState != state {
		c.bot.SendMessage(
			telegramId,
			"Обнаружен неверный OAuth state, попробуйте ещё раз",
			&gotgbot.SendMessageOpts{},
		)
		return fmt.Errorf("invalid state")
	}

	tok, err := c.deps.Splitwise.GetOAuthToken(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to get oauth token: %w", err)
	}

	err = c.deps.TokensDb.PutToken(ctx, &tokensdb.Token{TelegramId: telegramId, Token: tok})
	if err != nil {
		return fmt.Errorf("failed to put token: %w", err)
	}

	user.Authorized = true
	user.SplitwiseOAuthState = ""
	user.State = stateReady

	_, err = c.deps.UsersDb.UpdateUser(context.Background(), user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	_, err = c.bot.SendMessage(telegramId, "Вы авторизованы", &gotgbot.SendMessageOpts{})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (c *client) start(b *gotgbot.Bot, ctx *ext.Context) error {
	var err error
	user := ctx.Data["user"].(*usersdb.User)

	if user == nil {
		user = &usersdb.User{
			TelegramId:       ctx.EffectiveUser.Id,
			State:            stateUnauthorized,
			Currency:         "",
			SplitwiseGroupId: -1,
			Authorized:       false,
		}
		err = c.deps.UsersDb.PutUser(context.Background(), user)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		b.SendMessage(ctx.EffectiveChat.Id, "Добро пожаловать!", &gotgbot.SendMessageOpts{})
	}

	b.SendMessage(
		ctx.EffectiveChat.Id,
		"Привет!\n"+
			"Твоё текущее состояние:\n"+
			c.makeUserProfileString(user)+"\n"+
			"Чтобы узнать больше, введи /help",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: &gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{{
						Text:   "Open App",
						WebApp: &gotgbot.WebAppInfo{Url: c.webAppUrl},
					}},
				},
			},
		},
	)

	return nil
}

func (c *client) help(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(
		b,
		"Список команд:\n"+
			"/start - начать работу с ботом\n"+
			"/authorize - авторизоваться в Splitwise\n"+
			"/set_group - выбрать группу\n"+
			"/set_currency - выбрать валюту\n"+
			"Когда твой профиль будет заполнен, присылай фотку чека и начнём работать ;)",
		&gotgbot.SendMessageOpts{},
	)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func newMessageFilter(msg *gotgbot.Message) bool {
	return true
}

func (c *client) authorize(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.Data["user"].(*usersdb.User)

	oauthState, err := makeState(user.TelegramId)
	if err != nil {
		return fmt.Errorf("failed to generate random state: %w", err)
	}
	oauthUrl := c.deps.Splitwise.GetOAuthUrl(oauthState)
	b.SendMessage(
		ctx.EffectiveChat.Id,
		"Нажмите кнопку ниже, чтобы авторизоваться в Splitwise",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: &gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{{Text: "Авторизоваться", Url: oauthUrl}},
				},
			},
		},
	)

	user.State = stateAwaitingOAuthCode
	user.SplitwiseOAuthState = oauthState

	ctx.Data["user_dirty"] = true

	return nil
}

func (c *client) newMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.Data["user"].(*usersdb.User)

	if !user.Authorized {
		b.SendMessage(ctx.EffectiveChat.Id, "Вы не авторизованы", &gotgbot.SendMessageOpts{})
		return nil
	}

	if user.State == stateReady {
		photos := ctx.EffectiveMessage.Photo
		if photos != nil {
			photoFile, err := b.GetFile(photos[len(photos)-1].FileId, &gotgbot.GetFileOpts{})
			if err != nil {
				return fmt.Errorf("failed to get file: %w", err)
			}

			photoUrl := photoFile.URL(b, &gotgbot.RequestOpts{})

			b.SendMessage(ctx.EffectiveChat.Id, "Обрабатываю чек...", &gotgbot.SendMessageOpts{})

			cheque, err := c.deps.Ocr.GetChequeTranscription(photoUrl)
			if err != nil {
				return fmt.Errorf("failed to get vision: %w", err)
			}

			textBytes, _ := json.MarshalIndent(cheque, "", "  ")

			_, err = b.SendMessage(
				ctx.EffectiveChat.Id,
				fmt.Sprintf("<pre><code class=\"language-json\">%s</code></pre>", string(textBytes)),
				&gotgbot.SendMessageOpts{
					ParseMode: "HTML",
				},
			)
			if err != nil {
				return fmt.Errorf("failed to send message: %w", err)
			}
		}
		if ctx.EffectiveMessage.Voice != nil {
			voiceText, err := c.getVoiceTranscription(ctx.EffectiveMessage)
			if err != nil {
				return fmt.Errorf("failed to get voice transcription: %w", err)
			}
			ctx.EffectiveMessage.Reply(b, voiceText, &gotgbot.SendMessageOpts{})
		}
	}

	return nil
}

func (c *client) getVoiceTranscription(msg *gotgbot.Message) (string, error) {
	voice := msg.Voice
	c.log.Debugw("voice mime type", "mime", voice.MimeType)
	voiceFile, err := c.bot.GetFile(voice.FileId, &gotgbot.GetFileOpts{})
	if err != nil {
		return "", fmt.Errorf("failed to get file: %w", err)
	}
	c.log.Debugw("got voice file", "file", voiceFile)
	voiceFileURL := voiceFile.URL(c.bot, &gotgbot.RequestOpts{})

	voiceFilePath, err := downloadFile(voiceFileURL)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	c.log.Debugw("downloaded voice file", "path", voiceFilePath)

	voiceText, err := c.deps.OpenAI.GetTranscription(voiceFilePath, "")
	if err != nil {
		return "", fmt.Errorf("failed to get transcription: %w", err)
	}
	return *voiceText, nil
}

func (c *client) setGroup(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.Data["user"].(*usersdb.User)

	if !user.Authorized {
		_, err := b.SendMessage(ctx.EffectiveChat.Id, "Вы не авторизованы", &gotgbot.SendMessageOpts{})
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	splitwiseInstance := ctx.Data["splitwise_instance"].(splitwise.Instance)

	groups, err := splitwiseInstance.GetGroups(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get groups: %w", err)
	}
	keyboardButtons := make([][]gotgbot.InlineKeyboardButton, 0)
	for _, group := range groups {
		keyboardButtons = append(keyboardButtons, []gotgbot.InlineKeyboardButton{{
			Text:         group.Name,
			CallbackData: "g" + fmt.Sprintf("%x", group.ID),
		}})
	}
	_, err = b.SendMessage(
		ctx.EffectiveChat.Id,
		"Выберите группу:",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: &gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: keyboardButtons,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func (c *client) callbackQuery(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.Data["user"].(*usersdb.User)

	if !user.Authorized {
		b.SendMessage(ctx.EffectiveChat.Id, "Вы не авторизованы", &gotgbot.SendMessageOpts{})
		return nil
	}

	splitwiseInstance := ctx.Data["splitwise_instance"].(splitwise.Instance)

	query := ctx.CallbackQuery

	if query.Data[0] == 'g' {
		var groupId uint64
		fmt.Sscanf(query.Data, "g%x", &groupId)
		group, err := splitwiseInstance.GetGroup(context.Background(), int(groupId))
		if err != nil {
			b.SendMessage(ctx.EffectiveChat.Id, "Неверный id группы", &gotgbot.SendMessageOpts{})
			c.log.Errorw("failed to get group", "error", err)
			return nil
		}

		user.SplitwiseGroupId = int64(group.ID)
		ctx.Data["user_dirty"] = true

		b.EditMessageText(
			"Группа выбрана!\nВаш профиль:\n"+c.makeUserProfileString(user),
			&gotgbot.EditMessageTextOpts{
				ChatId:      ctx.EffectiveChat.Id,
				MessageId:   query.Message.MessageId,
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{},
			},
		)
	}

	return nil
}

func (c *client) setCurrency(b *gotgbot.Bot, ctx *ext.Context) error {
	var err error

	user := ctx.Data["user"].(*usersdb.User)

	if !user.Authorized {
		_, err = b.SendMessage(ctx.EffectiveChat.Id, "Вы не авторизованы", &gotgbot.SendMessageOpts{})
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	var currency string
	n, err := fmt.Sscanf(ctx.Message.Text, "/set_currency %s", &currency)
	if err != nil || n != 1 {
		b.SendMessage(ctx.EffectiveChat.Id, "Неверный формат", &gotgbot.SendMessageOpts{})
		return nil
	}

	splitwiseInstance := ctx.Data["splitwise_instance"].(splitwise.Instance)

	currencies, err := splitwiseInstance.GetCurrencies(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get currencies: %w", err)
	}
	currencyExists := false
	for _, cur := range currencies {
		if cur.CurrencyCode == currency {
			currencyExists = true
			break
		}
	}

	if !currencyExists {
		b.SendMessage(ctx.EffectiveChat.Id, "Неверная валюта", &gotgbot.SendMessageOpts{})
		return nil
	}

	user.Currency = currency
	ctx.Data["user_dirty"] = true

	b.SendMessage(ctx.EffectiveChat.Id, "Валюта выбрана", &gotgbot.SendMessageOpts{})
	b.SendMessage(ctx.EffectiveChat.Id, c.makeUserProfileString(user), &gotgbot.SendMessageOpts{})

	return nil
}
