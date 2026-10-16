import asyncio
from datetime import datetime, timezone
import os
from pathlib import Path
import re
from urllib.parse import parse_qs, quote, unquote, urlparse

import cadence


tools_registry = cadence.Registry()


def resolve_output_dir() -> Path:
    output_dir = Path(os.environ.get("AUTO_RESEARCH_OUTPUT_DIR", "research_reports")).expanduser().resolve()
    allowed_roots = [Path.cwd().resolve()]
    try:
        allowed_roots.append(Path.home().resolve())
    except RuntimeError:
        pass
    if not any(output_dir.is_relative_to(root) for root in allowed_roots):
        raise ValueError(
            f"AUTO_RESEARCH_OUTPUT_DIR must be inside one of {[str(r) for r in allowed_roots]}, got {output_dir}"
        )
    return output_dir


@tools_registry.activity(name="search_research_sources")
async def search_research_sources(query: str) -> str:
    """Return simulated source results scoped to a research question."""
    await asyncio.sleep(1)
    encoded_query = quote(query, safe="")
    return (
        f"Search results for '{query}':\n"
        f"1. research.example/overview?topic={encoded_query} — Overview and current evidence\n"
        f"2. research.example/challenges?topic={encoded_query} — Challenges and limitations\n"
        f"3. research.example/outlook?topic={encoded_query} — Recent developments and outlook"
    )


@tools_registry.activity(name="read_research_source")
async def read_research_source(url: str) -> str:
    """Return a simulated source brief for the topic encoded in a result URL."""
    await asyncio.sleep(1)
    parsed = urlparse(url)
    topic = unquote(parse_qs(parsed.query).get("topic", ["the requested topic"])[0])
    angle = parsed.path.rsplit("/", 1)[-1].replace("-", " ")
    return (
        f"Research brief: {topic} — {angle}\n"
        f"This source is specifically about '{topic}'. Analyze established facts, "
        "important evidence, practical implications, limitations, and future outlook. "
        "Distinguish established knowledge from uncertain or emerging claims."
    )


@tools_registry.activity(name="log_research_progress")
async def log_research_progress(message: str) -> str:
    """Print progress once as a durable activity rather than during workflow replay."""
    print(message, flush=True)
    return message


@tools_registry.activity(name="publish_research_report")
async def publish_research_report(report: str, topic: str) -> str:
    """Write the approved report to the configured local output directory."""
    output_dir = resolve_output_dir()
    output_dir.mkdir(parents=True, exist_ok=True)

    slug = re.sub(r"[^a-z0-9]+", "-", topic.lower()).strip("-")[:60]
    timestamp = datetime.now(timezone.utc).strftime("%Y%m%d-%H%M%S")
    output_path = (output_dir / f"{slug}-{timestamp}.md").resolve()
    output_path.write_text(report.strip() + "\n", encoding="utf-8")

    return f"Approved report saved to {output_path}"
