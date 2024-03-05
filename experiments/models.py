from langchain_openai import ChatOpenAI
from config.openai_config import *
from langchain_core.runnables import ConfigurableField

chat_model = ChatOpenAI(
    api_key=OPENAI_API_KEY,
    organization=OPENAI_ORGANIZATION_ID,
    max_tokens=MAX_TOKENS_FOR_RESPONSE,
    timeout=TIMEOUT,
    max_retries=MAX_RETRIES,
    seed=SEED,
    temperature=TEMPERATURE
).configurable_fields(
    temperature=ConfigurableField(id="temperature"),
    model_kwargs=ConfigurableField(id="model_kwargs"),
)
