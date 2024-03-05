from chains import picture_recognition_chain

CONFIGS = {
    "picture_recognition": {
        "image_res" : 1.0
    }
}


async def picture_recognition(image_path) -> list[dict[str, str]]:
    config = CONFIGS['picture_recognition']
    qna_sequence = await picture_recognition_chain.ainvoke(image_path, config=config)
    return qna_sequence
