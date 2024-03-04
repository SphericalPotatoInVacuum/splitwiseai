from inspect import cleandoc
from langchain.prompts.chat import ChatPromptTemplate
from langchain_core.messages import BaseMessage, HumanMessage, SystemMessage
from langchain_core.runnables import RunnableLambda
from langchain_openai import ChatOpenAI


system_template = """ 
"""

def _get_messages_from_image(base64_image) -> list[BaseMessage]:
    return [
        SystemMessage(content=cleandoc(system_template)),
        HumanMessage(
            content=[{
                "type": "image_url",
                "image_url": {
                    "url": f"data:image/jpeg;base64,{base64_image}"
                }
            }],
        ),
    ]



picture_recognition_template = RunnableLambda(_get_messages_from_image)