
# Debtease — Stress & Shock Simulation Domain (v1)

**Status:** v1 (Foundational)

---
## 1. Domain Overview

### 1.1 What this domain does

The **Stress & Shock Simulation domain** evaluates **how a user’s debt system behaves under adverse but realistic conditions**, without giving advice or recommendations.

It answers one core question:

> **“Where does this debt setup break when life doesn’t go as planned?”**

This domain:

* does **not** optimize repayments
* does **not** recommend actions
* does **not** predict default probability

It strictly identifies:

* fragility
* breaking points
* failure order
* loss of flexibility

---

### 1.2 What this domain explicitly does NOT do

❌ No optimization
❌ No planning
❌ No AI / ML
❌ No probabilistic outcomes
❌ No direct user-facing advice

Those belong to **Decision Intelligence** and **Debt Planner** domains.

---

## 2. Why This Domain Exists Separately

### 2.1 Separation from Repayment Domain

| Repayment Domain   | Stress Domain         |
| ------------------ | --------------------- |
| Assumes stability  | Assumes disruption    |
| Optimizes outcomes | Explores constraints  |
| Debt-level math    | System-level behavior |
| “Best strategy”    | “Failure threshold”   |

Mixing stress logic into repayment would:

* pollute optimization logic
* make assumptions unclear
* break explainability

---

### 2.2 Separation from Classification Domain

**Classification** answers:

> “What kind of risk does this debt carry by nature?”

**Stress** answers:

> “When does the system fail and what fails first?”

Risk type ≠ failure point.
Both are needed.

---

## 3. Position in Overall Debtease Architecture

```
Financial State
    ↓
Debt Classification
    ↓
Repayment Engine
    ↓
Stress & Shock Simulation   ← THIS DOMAIN
    ↓
Decision Intelligence
    ↓
Debt Planner
    ↓
UX
```

Stress is **never exposed directly** to users.
It only informs **which insight should be surfaced**.

---

## 4. User Types & How Stress Is Utilized

### 4.1 User Type A — Users WITH existing debts

#### Scenario A1: “Am I already in trouble?”

* Stress runs automatically
* User sees a calm status:

  * Resilient
  * Sensitive
  * Fragile
* No numbers by default

#### Scenario A2: “Why do small EMIs feel heavy?”

* Stress detects:

  * cashflow fragmentation
  * early failure of short-term debts
* Smart Debts explains impact without panic

#### Scenario A3: “Should I add one more EMI?”

* Stress evaluates marginal fragility
* Output is comparative, not absolute

---

### 4.2 User Type B — Users WITHOUT debts (explorers)

#### Scenario B1: “What happens if I take my first loan?”

* Stress evaluates rigidity and lock-in
* Output avoids risk language

#### Scenario B2: “BNPL looks harmless”

* Stress models stacking effect
* Explains structural risk, not failure

**Important:**
Stress outputs are always hypothetical and softly framed for this user type.

---

## 5. Debt Coverage in v1

### Included

* Home loans
* Personal loans
* Vehicle loans
* Consumer durable loans
* BNPL
* 0% EMI
* Personal line of credit

### Explicitly excluded (v1)

* Credit cards
* Overdrafts
* Informal loans

---

## 6. Mathematical Model (v1)

### 6.1 Core Variables

```
Disposable Income = Income − Fixed Expenses
Total EMI = Σ EMI_i
Flexibility Buffer = Disposable Income − Total EMI
```

This buffer is the **primary state variable**.

---

### 6.2 Debt Representation (post-repayment)

Each debt is represented as:

* EMI
* Remaining tenure
* Debt category
* Rigidity (deferrable vs non-deferrable)
* Failure priority

Stress never recalculates EMI logic.

---

## 7. Shock Scenarios (Deterministic)

### 7.1 Income Shock

```
Income' = Income × α
α ∈ {0.9, 0.8}
```

---

### 7.2 Interest Rate Shock (selective)

```
EMI_i' = EMI_i × (1 + β_i)
```

Where:

* β_i > 0 only for floating unsecured loans
* β_i = 0 for BNPL / 0% EMI / fixed loans

---

### 7.3 Expense Creep (optional later)

```
FixedExpenses' = FixedExpenses + Δ
```

---

## 8. Stress Computation Pipeline

### Step 1: Recompute Buffer

```
Flexibility' = Income' − FixedExpenses' − Σ EMI_i'
```

---

### Step 2: Survivability Band (Not Precise Months)

Instead of exact savings modeling, output bands:

* Immediate stress (0–1 month)
* Short survivability (1–3 months)
* Moderate survivability (3–6 months)
* Stable (>6 months)

This avoids false precision.

---

### Step 3: Failure Ordering

Debts are dropped in fixed priority order:

1. BNPL / 0% EMI
2. Short-term unsecured loans
3. Long-term unsecured loans
4. Secured loans

Algorithm:

* Consume deficit by removing EMIs in order
* First removed debt = primary failure

---

## 9. Stress Metrics (Outputs)

### 9.1 Stress Outcome Object

* Stress level (Resilient / Sensitive / Fragile)
* Flexibility after shock
* Survivability band
* Failed debt categories
* Key stress drivers (max 3)

---

### 9.2 Stress Level Classification

Policy-based thresholds:

* Positive buffer above threshold → Resilient
* Near-zero buffer → Sensitive
* Negative buffer → Fragile

Thresholds are policy-controlled, not hardcoded.

---

## 10. Explainability & Reasoning

### 10.1 Outcome vs Reasoning

* **Outcome** is always shown
* **Reasoning** is expandable
* **Evidence** (graphs) is optional

---

### 10.2 Reasoning Structure

Each outcome maps to:

* 2–3 deterministic drivers
* No probabilities
* No formulas shown to users

Example drivers:

* Low flexibility buffer
* High short-term rigidity
* Unsecured exposure

---

### 10.3 Natural Language (No AI)

Language is:

* template-based
* rule-driven
* deterministic
* auditable

No LLM usage in v1.

---

## 11. UX Exposure Rules (Critical)

### Must

* Stress is contextual
* Stress is calm
* Stress is optional to explore deeper

### Must NOT

* Standalone stress calculator
* Red failure charts
* “Default” language
* Fear-based warnings

Stress is always wrapped inside **Smart Debts**.

---

## 12. Invariants (Non-Negotiable)

* Same input → same output
* Stress never invents debt
* Stress never gives advice
* Stress never hides assumptions
* Stress never predicts certainty

---

## 13. Engineering Structure (Recommended)

```
domain/stress/v1
├── engine.go
├── scenarios.go
├── failure.go
├── metrics.go
├── policy.go
├── invariants.go
├── types.go
└── test/
    ├── engine_test.go
    ├── failure_test.go
    ├── golden_test.go
```

Golden tests are mandatory for trust.

---

## 14. Success Criteria (v1)

This domain is successful if:

* Users understand *why* something feels risky
* Users are not scared or overwhelmed
* Decision engine consistently selects correct insights
* Outputs remain stable across versions

---

## 15. Known Limitations (Accepted in v1)

* No savings modeling
* No probabilistic outcomes
* No behavioral predictions
* No real-time bank data

These are **intentional**.

---

## 16. Core Philosophy (Lock This)

> **Stress & Shock exists to reveal fragility, not to induce fear.
> It is an internal lens, not a user toy.**
