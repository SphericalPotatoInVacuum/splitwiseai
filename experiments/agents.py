from chains import picture_recognition_chain, voice_extraction_chain

CONFIGS = {
    "picture_recognition": {
        "max_size" : 1020,
        "image_format" : "PNG",
        "image_percent_quality" : 1.0
    }
}


async def picture_recognition(image_path) -> list[dict[str, str]]:
    config = CONFIGS['picture_recognition']
    response = await picture_recognition_chain.ainvoke(image_path, config=config)
    return response


async def voice_extraction(order_description) -> list[dict[str, str]]:
    response = await voice_extraction_chain.ainvoke(order_description)
    return response

