from langchain.schema import BaseOutputParser
import json

class JSONParser(BaseOutputParser):
    def parse(self, text: str):
        try:
            data = json.loads(text)
            return data
        except json.JSONDecodeError:
            raise ValueError("Invalid JSON format")

json_parser = JSONParser()
