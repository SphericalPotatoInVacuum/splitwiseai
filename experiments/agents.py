from chains import picture_recognition_chain


async def picture_recognition(image_path, image_resolution=1.0) -> list[dict[str, str]]:
    qna_sequence = await picture_recognition_chain.ainvoke({"image_path" : image_path, "image_res" : image_resolution})
    return qna_sequence
