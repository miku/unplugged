# agentic models

* pretraining (10T tokens)
* reasoning RL (1M problems)
* post-training (PT), RLHF (100k problems)

Finetuning, domain specific data (10k-100k problems)

Often, the bottleneck is "evals", i.e.

* Accuracy
* Recall
* F1 score
* Coherence
* Perplexity
* BLEU
* ROUGE
* Latency
* Toxicity

Also: LLM-as-a-judge

## training notes

* midtraining: Data mix changes shifts: eg more scientific, coding, multilingual data

## cost

* SOTA model, llama 3, 400B, 70 days, 16K H100, $50-80M; 2k JFK-LHR (moving 1M
  people over the atlantic, quickly)

## PT

* classic PT: usefulness of real-world tasks; 5k-500k problems (2022)
* reasoning: hard task with verifiable answer (2024)
* SFT, RL

> data of desired behaviors is what we want but scarce and expensive;
> pretraining data scales but is not what we want

data collection can be slow; "alpaca" to scale question and answer format
(https://crfm.stanford.edu/2023/03/13/alpaca.html)

> 52K instruction-following demonstrations

pretraining and "high-quality instruction-following data."

> synthetic data generation

Rejection sampling based on verifierds; let LLM generate answers, keep only
verified/correct answers.

SFT:

* intruction following
* desired format, style
* tool use (e.g. for agents, kimi 2, xLAM)
* early reasoning
* anything with good input, output pairs


## tool use PT

> Eg Kimi K2 uses LLM simulated user & tools, and rubric based rejection
> sampling to build data for agentic tool use

Kimi K2 (2026):

> The development of Large Language Models (LLMs) is undergoing a profound
> paradigm shift towards Agentic Intelligence – the capabilities for models to
> autonomously perceive, plan, reason, and act within complex and dynamic
> environments. This transition marks a departure from static imitation
> learning towards models that actively learn through interactions, acquire new
> skills beyond their training distribution, and adapt behavior through
> experiences [64].  It is believed that this approach allows an AI agent to go
> beyond the limitation of static human-generated data, and acquire superhuman
> capabilities through its own exploration and exploitation. Agentic
> intelligence is thus rapidly emerging as a defining capability for the next
> generation of foundation models, with wide-ranging implications across tool
> use, software development, and real-world autonomy

agentic interactions at scale

> We introduce a large-scale agentic data synthesis pipeline that
> systematically generates tool-use demonstrations via simulated and real-world
> environments. This system constructs diverse tools, agents, tasks, and
> trajectories to create high-fidelity, verifiably correct agentic interactions
> at scale.

SFT is quick; "You learn quickly from SFT! ~ 10k are sufficient for learning style & IF"

Preferences: RLHF (preference between two choices)

## refs

* [https://rdi.berkeley.edu/agentic-ai/slides/lecture1.pdf](https://rdi.berkeley.edu/agentic-ai/slides/lecture1.pdf)
