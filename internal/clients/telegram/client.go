package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"splitwiseai/internal/clients/db/tokensdb"
	"splitwiseai/internal/clients/db/usersdb"
	"splitwiseai/internal/clients/ocr"
	"splitwiseai/internal/clients/openai"
	"splitwiseai/internal/clients/splitwise"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"go.uber.org/zap"
)

type Config struct {
	TelegramBotToken string `env:"TELEGRAM_BOT_TOKEN"`
	SplitwiseCfg     splitwise.Config
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
}

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
	}

	dispatcher.AddHandlerToGroup(handlers.NewMessage(filterUserUpdates, client.preprocess), -1)

	dispatcher.AddHandler(handlers.NewCommand("start", client.start))
	dispatcher.AddHandler(handlers.NewCommand("help", client.help))
	dispatcher.AddHandler(handlers.NewCommand("authorize", client.authorize))
	dispatcher.AddHandler(handlers.NewCommand("set_group", client.setGroup))
	dispatcher.AddHandler(handlers.NewCommand("set_currency", client.setCurrency))
	dispatcher.AddHandler(handlers.NewCommand("get_groups", client.getGroups))

	dispatcher.AddHandler(handlers.NewMessage(newMessageFilter, client.newMessage))

	dispatcher.AddHandlerToGroup(handlers.NewMessage(filterUserUpdates, client.postprocess), 1000)

	return client, nil
}

func filterUserUpdates(message *gotgbot.Message) bool {
	return message != nil && message.From != nil && !message.From.IsBot
}

func (c *client) preprocess(b *gotgbot.Bot, ctx *ext.Context) error {
	user, err := c.deps.UsersDb.GetUser(context.Background(), ctx.EffectiveUser.Id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	ctx.Data["user"] = user
	ctx.Data["user_dirty"] = false

	if user == nil {
		return nil
	}

	if !user.Authorized {
		return nil
	}

	splitwiseInstance, ok := c.deps.Splitwise.GetInstance(user.TelegramId)
	if !ok {
		token, err := c.deps.TokensDb.GetToken(context.Background(), user.TelegramId)
		if err != nil {
			err = fmt.Errorf("failed to get token: %w", err)
		} else {
			if token == (tokensdb.Token{}) {
				err = fmt.Errorf("token not found")
			} else {
				splitwiseInstance, err = c.deps.Splitwise.AddInstanceFromOAuthToken(context.Background(), user.TelegramId, token.Token)
				if err != nil {
					err = fmt.Errorf("failed to add instance from oauth token: %w", err)
				}
			}
		}
		if err != nil {
			b.SendMessage(ctx.EffectiveChat.Id, "Авторизуйтесь", &gotgbot.SendMessageOpts{})
			return fmt.Errorf("user is authorized but splitwise instance couldn't be found: %w", err)
		}
	}
	ctx.Data["splitwise_instance"] = splitwiseInstance

	return nil
}

func (c *client) postprocess(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.Data["user"].(*usersdb.User)
	userDirty := ctx.Data["user_dirty"].(bool)

	if !userDirty {
		return nil
	}

	if user.Authorized && user.Currency != "" {
		user.State = usersdb.Ready.String()
	} else {
		user.State = usersdb.IncompleteProfile.String()
	}

	_, err := c.deps.UsersDb.UpdateUser(context.Background(), user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (c *client) makeUserProfileString(user *usersdb.User) string {
	var authorizedStr, groupString, currencyStr string

	if user.Authorized {
		authorizedStr = "✔️"
	} else {
		authorizedStr = "✖️"
	}

	groupString = fmt.Sprint(user.SplitwiseGroupId)

	if user.Currency != "" {
		currencyStr = user.Currency
	} else {
		currencyStr = "✖️"
	}

	return fmt.Sprintf(
		"Authorized: %s\nSelected group: %s\nSelected currency: %s\n",
		authorizedStr, groupString, currencyStr,
	)
}

func (c *client) HandleUpdate(update *gotgbot.Update) error {
	c.log.Debugw("handling update", "update", update)
	return c.dispatcher.ProcessUpdate(c.bot, update, map[string]interface{}{})
}

func (c *client) Auth(authUrl string) error {
	code, state, err := parseOAuth2RedirectURL(authUrl)
	if err != nil {
		return fmt.Errorf("failed to parse oauth2 redirect url: %w", err)
	}
	c.log.Debugw("handling splitwise callback", "state", state, "code", code)
	telegramId, _, err := parseState(state)
	if err != nil {
		return fmt.Errorf("failed to parse state: %w", err)
	}
	user, err := c.deps.UsersDb.GetUser(context.Background(), telegramId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user.SplitwiseOAuthState != state {
		c.bot.SendMessage(telegramId, "Неверный state", &gotgbot.SendMessageOpts{})
		return fmt.Errorf("invalid state")
	}
	tok, err := c.deps.Splitwise.GetOAuthToken(context.Background(), code)
	if err != nil {
		return fmt.Errorf("failed to get oauth token: %w", err)
	}
	err = c.deps.TokensDb.PutToken(context.Background(), &tokensdb.Token{TelegramId: telegramId, Token: tok})
	if err != nil {
		return fmt.Errorf("failed to put token: %w", err)
	}
	user.Authorized = true
	user.SplitwiseOAuthState = ""
	if user.Currency != "" {
		user.State = usersdb.Ready.String()
	} else {
		user.State = usersdb.IncompleteProfile.String()
	}
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
			TelegramId: ctx.EffectiveUser.Id,
			State:      usersdb.IncompleteProfile.String(),
			Authorized: false,
		}
		err = c.deps.UsersDb.CreateUser(context.Background(), user)
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
		&gotgbot.SendMessageOpts{},
	)

	return nil
}

func (c *client) help(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(
		b,
		"Список команд:\n"+
			"/start - начать работу с ботом\n"+
			"/authorize - авторизоваться в Splitwise\n"+
			"/get_groups - получить список групп\n"+
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

	user.State = usersdb.AwaitingOAuthCode.String()
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

	if user.State == usersdb.Ready.String() {
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
			voice := ctx.EffectiveMessage.Voice
			c.log.Debugw("voice mime type", "mime", voice.MimeType)
			voiceFile, err := b.GetFile(voice.FileId, &gotgbot.GetFileOpts{})
			if err != nil {
				return fmt.Errorf("failed to get file: %w", err)
			}
			c.log.Debugw("got voice file", "file", voiceFile)
			voiceFileURL := voiceFile.URL(b, &gotgbot.RequestOpts{})

			voiceFilePath, err := downloadFile(voiceFileURL)
			if err != nil {
				return fmt.Errorf("failed to download file: %w", err)
			}
			c.log.Debugw("downloaded voice file", "path", voiceFilePath)

			voiceText, err := c.deps.OpenAI.GetTranscription(voiceFilePath, "")
			if err != nil {
				return fmt.Errorf("failed to get transcription: %w", err)
			}
			ctx.EffectiveMessage.Reply(b, voiceText, &gotgbot.SendMessageOpts{})
		}
	}

	return nil
}

func (c *client) getGroups(b *gotgbot.Bot, ctx *ext.Context) error {
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
	if len(groups) == 0 {
		_, err = b.SendMessage(ctx.EffectiveChat.Id, "Нет групп", &gotgbot.SendMessageOpts{})
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	}
	groupsString := ""
	for _, group := range groups {
		groupsString += fmt.Sprintf("%s: %d\n", group.Name, group.ID)
	}
	_, err = b.SendMessage(ctx.EffectiveChat.Id, groupsString, &gotgbot.SendMessageOpts{})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func (c *client) setGroup(b *gotgbot.Bot, ctx *ext.Context) error {
	var err error

	user := ctx.Data["user"].(*usersdb.User)

	if !user.Authorized {
		_, err = b.SendMessage(ctx.EffectiveChat.Id, "Вы не авторизованы", &gotgbot.SendMessageOpts{})
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	var groupId uint64
	n, err := fmt.Sscanf(ctx.Message.Text, "/set_group %d", &groupId)
	if err != nil || n != 1 {
		b.SendMessage(ctx.EffectiveChat.Id, "Неверный формат", &gotgbot.SendMessageOpts{})
		return nil
	}

	splitwiseInstance := ctx.Data["splitwise_instance"].(splitwise.Instance)

	group, err := splitwiseInstance.GetGroup(context.Background(), int(groupId))
	if err != nil {
		b.SendMessage(ctx.EffectiveChat.Id, "Неверный id группы", &gotgbot.SendMessageOpts{})
		return nil
	}

	user.SplitwiseGroupId = uint64(group.ID)
	ctx.Data["user_dirty"] = true

	b.SendMessage(ctx.EffectiveChat.Id, "Группа выбрана", &gotgbot.SendMessageOpts{})
	b.SendMessage(ctx.EffectiveChat.Id, c.makeUserProfileString(user), &gotgbot.SendMessageOpts{})

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
