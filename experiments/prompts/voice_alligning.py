from inspect import cleandoc
from langchain.prompts.chat import ChatPromptTemplate

system_template = """
"""

human_template = """

"""
    

voice_alligning_template = ChatPromptTemplate.from_messages([
    ("system", cleandoc(system_template)),
    ("human", cleandoc(human_template)),
])
