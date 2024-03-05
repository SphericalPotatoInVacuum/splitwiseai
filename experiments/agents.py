from chains import picture_recognition_chain

CONFIGS = {
    "picture_recognition": {
        "max_size" : 1020,
        "image_format" : "PNG",
        "image_percent_quality" : 1.0
    }
}


async def picture_recognition(image_path) -> list[dict[str, str]]:
    config = CONFIGS['picture_recognition']
    qna_sequence = await picture_recognition_chain.ainvoke(image_path, config=config)
    return qna_sequence
