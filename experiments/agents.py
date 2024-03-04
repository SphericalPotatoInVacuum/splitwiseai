from chains import picture_recognition_chain
from langchain.callbacks import get_openai_callback
from utils import encode_image
from PIL import Image


async def picture_recognition(image_path) -> list[dict[str, str]]:
    base64_image = encode_image(image_path)
    qna_sequence = await picture_recognition_chain.ainvoke(base64_image)
    return qna_sequence
