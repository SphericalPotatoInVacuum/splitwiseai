from config.openai_config import GPT_3_NAME, GPT_4_VISION_NAME, GPT_4_NAME
from prompts.picture_recognition import picture_recognition_template
from models import chat_model 
from parsers import json_parser

picture_recognition_chain = picture_recognition_template | chat_model.bind(model=GPT_4_VISION_NAME) | json_parser
voice_alligning_chain = picture_recognition_template | chat_model.bind(model=GPT_4_NAME) | json_parser
voice_extraction_chain = picture_recognition_template | chat_model.bind(model=GPT_4_NAME) | json_parser
names_alligning_chain = picture_recognition_template | chat_model.bind(model=GPT_3_NAME) | json_parser
