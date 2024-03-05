from chains import picture_recognition_chain
from utils import encode_image


async def picture_recognition(image_path) -> list[dict[str, str]]:
    base64_image = encode_image(image_path)
    qna_sequence = await picture_recognition_chain.ainvoke(base64_image)
    return qna_sequence
