from chains import *

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
    response = await voice_extraction_chain.ainvoke({"description":order_description})
    return response


async def voice_aligning(guests_description, system_order) -> list[dict[str, str]]:
    short_system_order = [
        {"id" : r["id"], "name" : r["name"], "count" : r["count"]} for r in system_order["items"]
    ]
    
    response = await voice_aligning_chain.ainvoke({
        "guests_description":guests_description,
        "system_order" : short_system_order
    })
    response = [{'id' : r['id'], 'users' : r['guests']} for r in response]
    return response
