"""A generated module for SmartModule functions

This module has been generated via dagger init and serves as a reference to
basic module structure as you get started with Dagger.

Two functions have been pre-created. You can modify, delete, or add to them,
as needed. They demonstrate usage of arguments and return types using simple
echo and grep commands. The functions can be called from the dagger CLI or
from one of the SDKs.

The first line in this comment block is a short description line and the
rest is a long description with more detail on the module's purpose or usage,
if appropriate. All modules should have a short description.
"""

import dagger
from dagger import dag, function, object_type

from langchain_core.messages import HumanMessage, SystemMessage
from langchain_openai import ChatOpenAI
from langgraph.graph import START, StateGraph, MessagesState
from langgraph.prebuilt import tools_condition, ToolNode

from . import tools


@object_type
class SmartModule:
    def _init_tools(self) -> list:
        """Initialize the tools for the LLM"""
        tools_funcs = []
        for t in dir(tools):
            if not t.startswith("tool_"):
                continue
            func = getattr(tools, t)
            if not callable(func):
                continue
            tools_funcs.append(getattr(tools, t))
        return tools_funcs

    @function
    async def ask(self, api_key: dagger.Secret, prompt: str) -> str:
        """Ask the LLM a prompt that involves a dagger module call"""
        model = ChatOpenAI(model="gpt-4o", api_key=await api_key.plaintext())

        tools = self._init_tools()
        model = model.bind_tools(tools)

        workflow = StateGraph(MessagesState)

        async def call_model(state: MessagesState):
            messages = state["messages"]
            response = await model.ainvoke(messages)
            return {"messages": [response]}

        # Define nodes: these do the work
        workflow.add_node("agent", call_model)
        workflow.add_node("tools", ToolNode(tools))

        # Define edges: these determine how the control flow moves
        workflow.add_edge(START, "agent")
        workflow.add_conditional_edges(
            "agent",
            # Loop to tools until the last message is not a tool call (in that case, route to END)
            tools_condition,
        )
        workflow.add_edge("tools", "agent")
        app = workflow.compile()

        inputs = {"messages": [HumanMessage(content=prompt)]}
        messages = await app.ainvoke(inputs)

        for m in messages['messages']:
            m.pretty_print()

        return messages['messages'][-1].content
