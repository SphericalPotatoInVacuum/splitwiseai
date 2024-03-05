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

def _image_encoder(image_path):
    with open(image_path, "rb") as image_file:
        return base64.b64encode(image_file.read()).decode('utf-8')

def _image_encoder_with_resolution(image_path, config):
    image_percent_resolution = config['image_res']
    
    with open(image_path, "rb") as image_file:
        img = Image.open(image_file)
        buffered = BytesIO()
        img.save(buffered, format="JPEG", quality=int(image_percent_resolution * 100))
        return base64.b64encode(buffered.getvalue()).decode('utf-8')


json_parser = JSONParser()
image_encoder = RunnableLambda(_image_encoder)
