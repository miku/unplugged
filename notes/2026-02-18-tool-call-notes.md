> https://claude.ai/chat/026e1d53-5d30-4ecd-9d0c-053b9299b116

## Training LLMs for Tool Use: A Technical Overview

### The Core Idea

Tool-use training teaches a model to emit **structured output** (a function
call) instead of — or interleaved with — natural language, and then to
correctly *consume* the result returned by the tool. This is typically done via
supervised fine-tuning (SFT) on curated examples, followed by reinforcement
learning from human/automated feedback (RLHF/RLAIF).

---

### The Data Format

At its heart, the training data is multi-turn conversation with special
"roles": `user`, `assistant`, and `tool` (or `function`/`observation`,
depending on the framework). Here's a representative example in JSON-like
pseudoformat:

```json [ { "role": "system", "content": "You are a helpful assistant with
access to the following tools:\n\n[{\"name\": \"bash\", \"description\": \"Run
a bash command and return stdout/stderr.\", \"parameters\": {\"command\":
{\"type\": \"string\"}}}]" }, { "role": "user", "content": "How many lines are
in /etc/passwd?" }, { "role": "assistant", "content": null, "tool_calls": [ {
"id": "call_01", "type": "function", "function": { "name": "bash", "arguments":
"{\"command\": \"wc -l /etc/passwd\"}" } } ] }, { "role": "tool",
"tool_call_id": "call_01", "content": "47 /etc/passwd" }, { "role":
"assistant", "content": "There are 47 lines in /etc/passwd." } ] ```

The model is trained to **predict the assistant turns only** — both the
`tool_calls` JSON and the final natural-language response. The `tool` role turn
(the observation) is treated as a fixed input, not predicted.

---

### The Training Pipeline

**1. Supervised Fine-Tuning (SFT)**

The base LLM is fine-tuned on thousands to millions of (context → tool call →
observation → response) trajectories. The loss is computed only on the
assistant tokens:

``` L = -Σ log P(assistant_token_i | all_preceding_tokens) ```

This teaches the model the *syntax and grammar* of tool invocation.

**2. Rejection Sampling / Self-Play Data Generation**

You can scale data cheaply by:
- Sampling many candidate tool calls from the model
- Executing them in a sandbox
- Keeping trajectories that produced correct outcomes (pass tests, return
  expected values, etc.)
- Filtering bad trajectories out

This is sometimes called **STaR (Self-Taught Reasoner)** or rejection sampling
fine-tuning (RFT). [https://arxiv.org/abs/2203.14465, 05/2022]

**3. RLHF / RLAIF / PPO / DPO**

Preference pairs are constructed:

``` Preferred:  model calls `grep -r "TODO" ./src` → gets real results → good
summary Rejected:   model hallucinates output instead of calling the tool ```

DPO (Direct Preference Optimization) or PPO with a reward model is then used to
push the model toward correct tool-use behavior. For code/bash, the reward
signal can be **fully automated** — you just check exit codes, test suite
results, or expected stdout.

**4. Process Reward Models (PRMs)**

For multi-step agentic tasks (e.g., a 10-step bash session), a PRM gives
intermediate step-level rewards rather than just terminal rewards,
significantly improving training signal quality.

---

### What Makes Bash/Terminal Training Data Special

Terminal tool use is interesting because:

- **The reward is mostly objective** — did the command succeed? Did the test
  pass? Was the file created?
- **Trajectories can be very long** — a bash session might span 20+ turns
- **Error recovery matters** — good training data includes examples of the
  model observing a non-zero exit code and correctly diagnosing and retrying

A multi-step bash training example might look like:

``` user: "Set up a Python virtualenv in ./env and install requests"

assistant: tool_call → {"command": "python3 -m venv ./env"} tool: "✓ (exit 0)"

assistant: tool_call → {"command": "./env/bin/pip install requests"} tool:
"Successfully installed requests-2.31.0"

assistant: "Done. The virtualenv is at ./env with requests installed." ```

And a recovery example:

``` assistant: tool_call → {"command": "pip install requests"} tool: "error:
externally-managed-environment (exit 1)"

assistant: tool_call → {"command": "pip install requests
--break-system-packages"} tool: "Successfully installed requests-2.31.0" ```

---

### Publicly Available Training Datasets

Yes, there are several openly available datasets. Here's a breakdown by
category:

**General Function/Tool Calling**

- **ToolBench / ToolLLM** (Qin et al., 2023) — ~16k real-world REST APIs, 126k
  tool-use trajectories using GPT-4 as a teacher. Probably the largest open
tool-use dataset. Available on HuggingFace: `ToolBench/ToolBench`.
- **APIBench** (Gorilla project, UC Berkeley) — covers HuggingFace, TorchHub,
  TensorHub APIs. Model + dataset at `ShishirPatil/gorilla`.
- **xLAM-function-calling-60k** — 60k function-calling examples from the
  Salesforce xLAM project, formatted for OpenAI-style tool use.
- **NousResearch/hermes-function-calling-v1** — ~11k examples in "Hermes"
  format, popular for fine-tuning Mistral/Llama models.
- **glaiveai/glaive-function-calling-v2** — ~110k synthetic function calling
  examples.

**Code / Terminal / Bash Specific**

- **InterCode** (Yang et al., 2023) — benchmark + trajectories for interactive
  bash and SQL coding tasks. Includes error-recovery examples. GitHub:
`princeton-nlp/intercode`.
- **SWE-bench** — GitHub issue → code fix trajectories. Not bash per se, but
  includes shell commands in agent traces. Many open agent traces are now
available from solvers like SWE-agent.
- **OSWorld / AndroidWorld** — computer-use trajectories (GUI + shell). OSWorld
  has open data.
- **SWE-agent trajectories** — Princeton's SWE-agent open-sources its bash
  agent traces on HuggingFace.
- **Shell Command Datasets** — `mzbac/shell-command-dataset`,
  `theblackcat102/shell-commands` — shorter single-command
instruction-following datasets.

**Agent / Multi-Step**

- **AgentInstruct** (Microsoft, 2024) — large synthetic agentic dataset,
  partially released, covering web browsing, file manipulation, and coding.
- **τ-bench** — tool-agent benchmark with open trajectories for retail/airline
  domains.
- **OpenAgentBench** — aggregates several agent task datasets.

---

### A Minimal Fine-Tuning Recipe (Practical)

If you wanted to fine-tune a Llama 3 or Mistral model for bash tool use today,
the typical recipe would be:

```
1. Start with a strong base (Llama 3.1 8B Instruct or Qwen2.5-Coder)
2. Format data as ChatML or Llama-3 tool-call format
3. SFT with LoRA (rank 64-128) on ~50k tool trajectories
   - Mix: ~40% successful trajectories, ~30% error+recovery, ~30% refusals/edge
     cases
4. Run rejection sampling: execute model's tool calls in a Docker sandbox, keep
correct ones, add to training set, repeat
5. Optional DPO pass on preferred vs. rejected trajectory pairs
6. Evaluate on InterCode-bash or custom held-out task suite ```

The key insight across all of this is that **bash/terminal tool use is one of
the easiest domains to get automated reward signals for**, which is why it's a
popular starting point for learning about agentic RL training — the environment
itself is the judge.
