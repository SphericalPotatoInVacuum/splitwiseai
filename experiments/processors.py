from langchain.schema import BaseOutputParser
from langchain_core.runnables import RunnableLambda
import json
from PIL import Image
from io import BytesIO
import base64

class JSONParser(BaseOutputParser):
    def parse(self, text: str):
        try:
            data = json.loads(text)
            return data
        except json.JSONDecodeError:
            raise ValueError("Invalid JSON format")


def _image_encoder(image_path, config):
    max_size_px = config['max_size']
    
    with open(image_path, "rb") as image_file:
        img = Image.open(image_file)

        if max_size_px:
            width, height = img.size
            max_dimension = max(width, height)
            if max_dimension > max_size_px:
                ratio = max_size_px / max_dimension
                new_width = int(width * ratio)
                new_height = int(height * ratio)
                img = img.resize((new_width, new_height), Image.BICUBIC)

        buffered = BytesIO()
        img.save(buffered, format="PNG", quality=100)
        return base64.b64encode(buffered.getvalue()).decode('utf-8')


json_parser = JSONParser()
image_encoder = RunnableLambda(_image_encoder)
