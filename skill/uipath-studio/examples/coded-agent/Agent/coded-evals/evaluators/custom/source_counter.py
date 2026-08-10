"""Custom evaluator that checks the number of sources returned."""

import json

from uipath.eval.evaluators import BaseEvaluator, BaseEvaluationCriteria, BaseEvaluatorConfig
from uipath.eval.models import AgentExecution, EvaluationResult, NumericEvaluationResult


class SourceCountCriteria(BaseEvaluationCriteria):
    """Evaluation criteria for the source count evaluator."""

    min_sources: int


class SourceCountConfig(BaseEvaluatorConfig[SourceCountCriteria]):
    """Configuration for the source count evaluator."""

    name: str = "SourceCountEvaluator"
    negated: bool = False
    default_evaluation_criteria: SourceCountCriteria = SourceCountCriteria(
        min_sources=1
    )


class SourceCountEvaluator(
    BaseEvaluator[SourceCountCriteria, SourceCountConfig, type(None)]
):
    """Evaluates whether the agent returned enough sources."""

    @classmethod
    def get_evaluator_id(cls) -> str:
        return "SourceCountEvaluator"

    async def evaluate(
        self,
        agent_execution: AgentExecution,
        evaluation_criteria: SourceCountCriteria,
    ) -> EvaluationResult:
        output = agent_execution.output
        if isinstance(output, str):
            output = json.loads(output)

        sources = output.get("sources", [])
        has_enough = len(sources) >= evaluation_criteria.min_sources

        if self.evaluator_config.negated:
            has_enough = not has_enough

        return NumericEvaluationResult(score=float(has_enough))
