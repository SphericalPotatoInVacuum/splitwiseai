import os
import openai
from dotenv import load_dotenv

load_dotenv()

OPENAI_API_KEY = os.environ['OPENAI_API_KEY']
OPENAI_ORGANIZATION_ID = os.environ['OPENAI_ORGANIZATION_ID']

MAX_TOKENS_FOR_RESPONSE = 3000
OPENAI_TIMEOUT = 120
MAX_RETRIES = 3
OPENAI_SEED = 42

GPT_3_NAME = "gpt-3.5-turbo-0125"
GPT_4_NAME = "gpt-4-0125-preview"
GPT_4_VISION_NAME = "gpt-4-vision-preview"
