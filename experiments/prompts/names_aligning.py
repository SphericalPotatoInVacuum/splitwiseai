from inspect import cleandoc
from langchain.prompts.chat import ChatPromptTemplate
from langchain_openai import ChatOpenAI

system_template = """

"""

human_template = """

"""
    

names_aligning_template = ChatPromptTemplate.from_messages([
    ("system", cleandoc(system_template)),
    ("human", cleandoc(human_template)),
])
