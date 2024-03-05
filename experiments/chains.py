from config.openai_config import GPT_3_NAME, GPT_4_VISION_NAME, GPT_4_NAME
from prompts.picture_recognition import picture_recognition_template
from prompts.voice_aligning import voice_aligning_template
from prompts.voice_extraction import voice_extraction_template
from prompts.names_aligning import names_aligning_template
from operator import itemgetter
from models import chat_model
from processors import json_parser, image_encoder

picture_recognition_chain = (
    {
        "image_path": itemgetter("image_path"),
        "image_res": itemgetter("image_res")
    } | image_encoder
    | picture_recognition_template
    | chat_model.bind(model=GPT_4_VISION_NAME)
    | json_parser
)
voice_aligning_chain = voice_aligning_template | chat_model.bind(
    model=GPT_4_NAME) | json_parser
voice_extraction_chain = voice_extraction_template | chat_model.bind(
    model=GPT_4_NAME) | json_parser
names_aligning_chain = names_aligning_template | chat_model.bind(
    model=GPT_3_NAME) | json_parser
