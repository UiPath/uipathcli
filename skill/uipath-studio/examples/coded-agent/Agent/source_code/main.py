"""UiPath Coded Agent — Research Assistant.

This agent processes natural language queries and returns structured answers.
Uses the UiPath Python SDK with tracing and evaluation support.
"""

import logging
from typing import Optional

from pydantic.dataclasses import dataclass
from uipath.eval.mocks import ExampleCall, mockable
from uipath.tracing import traced

logger = logging.getLogger(__name__)


@dataclass
class ResearchInput:
    """Agent input schema."""

    query: str
    max_results: Optional[int] = 5


@dataclass
class ResearchOutput:
    """Agent output schema."""

    answer: str
    sources: list[str]


# Example calls for evaluation simulation
SEARCH_EXAMPLES = [
    ExampleCall(
        id="search-example",
        input='{"query": "What is UiPath?"}',
        output='{"results": [{"title": "UiPath", "url": "https://uipath.com", "snippet": "Enterprise automation platform"}]}',
    )
]


@traced()
@mockable(example_calls=SEARCH_EXAMPLES)
async def search_web(query: str) -> dict:
    """Search the web for information.

    In production, this calls UiPath GenAI Activities web search.
    During evals, returns mock data from SEARCH_EXAMPLES.
    """
    # This would be replaced with actual UiPath tool call
    return {"results": []}


@traced(name="format_answer")
def format_answer(query: str, search_results: list[dict]) -> ResearchOutput:
    """Format search results into a structured answer."""
    sources = [r.get("url", "") for r in search_results if r.get("url")]
    snippets = [r.get("snippet", "") for r in search_results if r.get("snippet")]
    answer = f"Results for '{query}': " + " | ".join(snippets) if snippets else "No results found."
    return ResearchOutput(answer=answer, sources=sources)


@traced()
async def main(input: ResearchInput) -> ResearchOutput:
    """Main agent entry point.

    Searches the web for the query and returns a formatted answer.
    """
    logger.info("Processing query: %s", input.query)

    search_data = await search_web(input.query)
    results = search_data.get("results", [])

    if input.max_results:
        results = results[: input.max_results]

    return format_answer(input.query, results)
