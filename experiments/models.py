from langchain_openai import ChatOpenAI
from config.openai_config import OPENAI_API_KEY, MAX_TOKENS_FOR_RESPONSE, OPENAI_ORGANIZATION_ID, OPENAI_TIMEOUT, MAX_RETRIES, OPENAI_SEED

from config.openai_config import *

chat_model = ChatOpenAI(
    api_key=OPENAI_API_KEY,
    organization=OPENAI_ORGANIZATION_ID,
    max_tokens=MAX_TOKENS_FOR_RESPONSE,
    timeout=TIMEOUT,
    max_retries=MAX_RETRIES,
    seed=OPENAI_SEED,
    seed=SEED,
    temperature=TEMPERATURE
)
