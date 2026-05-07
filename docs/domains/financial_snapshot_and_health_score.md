# Debtease — Financial Snapshot, Schema Evolution & Financial Health Score Design

**Document Version:** 1.0
**Date:** 2026-03-11
**Owner:** Product + Platform
**Scope:** Financial state modeling, onboarding persistence, schema changes, and the Financial Health Score system.

---

# 1. Purpose

This document defines the **core financial state architecture of Debtease**.

It introduces:

* The **Financial Snapshot model**
* Required **database schema changes**
* The **Financial Health Score system**
* The relationship between **classification**, **stress simulation**, and **health scoring**

The goal is to create a **single financial state engine** that powers:

* onboarding analysis
* logged-in dashboards
* loan simulations
* debt management
* future financial insights

---

# 2. Core Architectural Principle

Debtease does not reason about **debts alone**.

It reasons about **financial context**.

Therefore the system architecture is:

```
Financial Snapshot (state)
        ↓
Debt Classification
        ↓
Stress Simulation
        ↓
Financial Health Score
        ↓
User Insights
```

The **financial snapshot** becomes the single source of truth for the user's financial state.

All engines operate on this state.

---

# 3. User Contexts Supported

The system supports two user contexts.

---

## 3.1 Pre-Signup Users (Session Mode)

Users can analyze their finances without creating an account.

Flow:

```
Landing Page
    ↓
Financial Snapshot Input
    ↓
Debt Health Analysis
    ↓
Loan Simulation (optional)
    ↓
Signup
```

Data is attached to a **session**.

---

## 3.2 Logged-In Users

Once a user signs up:

```
Login
    ↓
System builds financial snapshot
    ↓
Debt Health Dashboard
    ↓
Debt Portfolio Tracking
    ↓
Loan Simulations
    ↓
Repayment Planning
```

Snapshots are now attached to **user_id**.

---

# 4. Financial Snapshot Model

## 4.1 Definition

A **financial snapshot** represents the user's financial state at a specific point in time.

It stores **raw financial inputs**, not derived analysis.

Derived outputs such as:

* debt health
* stress levels
* health score

must always be **recomputed from snapshot data**.

---

## 4.2 Snapshot Lifecycle

Snapshots are created when financial state changes:

* onboarding input submitted
* income updated
* expenses updated
* debt added
* debt closed
* payment recorded

Snapshots are **event-driven**, not time-series sampled.

---

# 5. Financial Snapshot Schema

Table:

```
financial_snapshots
```

### Core fields

| Field           | Type            | Description                 |
| --------------- | --------------- | --------------------------- |
| id              | UUID            | Primary key                 |
| user_id         | UUID (nullable) | Linked user                 |
| session_id      | UUID (nullable) | Linked session              |
| snapshot_type   | TEXT            | onboarding / auto_generated |
| snapshot_source | TEXT            | user_input / calculated     |

---

### Income

| Field          | Type    |
| -------------- | ------- |
| monthly_income | NUMERIC |
| other_income   | NUMERIC |

---

### Expenses

| Field             | Type    |
| ----------------- | ------- |
| fixed_expenses    | NUMERIC |
| variable_expenses | NUMERIC |

---

### Assets

| Field             | Type    |
| ----------------- | ------- |
| savings_cash      | NUMERIC |
| investments_value | NUMERIC |

---

### Debt aggregates

| Field              | Type    |
| ------------------ | ------- |
| total_monthly_emi  | NUMERIC |
| total_debt_balance | NUMERIC |

---

### Household

| Field            | Type    |
| ---------------- | ------- |
| dependents_count | INTEGER |

---

### Metadata

| Field            | Type      |
| ---------------- | --------- |
| trigger_event    | TEXT      |
| snapshot_version | INTEGER   |
| extra_data       | JSONB     |
| created_at       | TIMESTAMP |

---

# 6. Why Financial Snapshots Exist

Snapshots provide three major capabilities.

---

## 6.1 Unified Engine Input

Both classification and stress engines require the same inputs:

```
income
fixed obligations
total EMI
```

Snapshots provide a consistent input state.

---

## 6.2 Financial Timeline

Snapshots allow reconstruction of the user's financial evolution.

Example timeline:

| Date | EMI | Income | Health    |
| ---- | --- | ------ | --------- |
| Jan  | 20k | 60k    | Fragile   |
| Mar  | 25k | 60k    | Sensitive |
| Jun  | 25k | 75k    | Stable    |

---

## 6.3 Engine Recomputability

Because snapshots store **raw state**, not conclusions, the system can:

```
recompute health metrics
re-evaluate policies
re-run simulations
```

without losing accuracy.

---

# 7. Schema Changes

## 7.1 Users Table

Users represent **identity**, not financial state.

No financial attributes are required here.

Example schema:

```
users
  id
  email
  name
  created_at
```

Financial state lives in `financial_snapshots`.

---

## 7.2 Debt Types Expansion

New debt types required for product scope:

```
personal_loan
home_loan
car_loan
consumer_emi
bnpl
credit_card
```

Migration example:

```
ALTER TYPE debt_type ADD VALUE 'car_loan';
ALTER TYPE debt_type ADD VALUE 'consumer_emi';
ALTER TYPE debt_type ADD VALUE 'bnpl';
ALTER TYPE debt_type ADD VALUE 'home_loan';
```

---

## 7.3 Sessions Table

Extend sessions to track onboarding funnels.

```
calculation_sessions
```

Add:

| Field       | Type            |
| ----------- | --------------- |
| user_id     | UUID (nullable) |
| entry_point | TEXT            |

Example values:

```
debt_health
loan_test
landing
```

---

## 7.4 Event Tracking

Existing table:

```
calculation_events
```

Add onboarding events:

```
landing_viewed
cta_clicked
financial_snapshot_submitted
debt_health_viewed
loan_simulation_started
loan_impact_viewed
signup_triggered
signup_completed
```

---

# 8. Loan Simulation Persistence

Loan simulations are **computed dynamically**.

For v1 they are **not stored**.

Simulation flow:

```
financial_snapshot
    +
hypothetical loan
    ↓
simulation engines
    ↓
impact result
```

Persistence can be added later if features require:

* saved scenarios
* simulation history
* comparison tools

---

# 9. Financial Health Score System

## 9.1 Purpose

Provide a simple user-facing signal representing the **structural health of the user's debt system**.

The score must remain:

* explainable
* deterministic
* policy-driven

---

## 9.2 Score Range

```
0 — 100
```

Higher score = stronger financial resilience.

---

## 9.3 Score Drivers

Score derives from structural metrics.

Primary metrics:

```
EMI Ratio
Shock Buffer
Debt Fragmentation
Unsecured Exposure
```

---

## 9.4 Score Penalties

Start with base score:

```
BASE_SCORE = 100
```

Apply penalties.

---

### EMI Ratio

| Range  | Penalty |
| ------ | ------- |
| <30%   | 0       |
| 30–45% | -10     |
| >45%   | -25     |

---

### Shock Buffer

| Range      | Penalty |
| ---------- | ------- |
| >6 months  | 0       |
| 3–6 months | -5      |
| 1–3 months | -15     |
| <1 month   | -30     |

---

### Fragmentation

| Debts | Penalty |
| ----- | ------- |
| ≤2    | 0       |
| 3–4   | -5      |
| ≥5    | -15     |

---

### Unsecured Exposure

| Ratio  | Penalty |
| ------ | ------- |
| <40%   | 0       |
| 40–60% | -5      |
| >60%   | -10     |

---

## 9.5 Final Score

```
score = BASE_SCORE − penalties
```

Clamp result:

```
0 ≤ score ≤ 100
```

---

# 10. Health Status Categories

Score maps to system health status.

| Score  | Status    |
| ------ | --------- |
| 80–100 | Stable    |
| 60–79  | Sensitive |
| <60    | Fragile   |

---

# 11. Relationship with Debt Classification

Per-loan classification remains separate.

Example:

```
Loan A → SAFE_TO_HOLD
Loan B → SAFE_TO_HOLD
Loan C → NEEDS_MONITORING
```

But system health evaluates **portfolio structure**.

Thus:

```
Loan Health → Component level
System Health → Portfolio level
```

---

# 12. Dashboard UX

Logged-in users see:

```
Financial Health Score
```

Example:

```
Score: 72
Status: Sensitive
```

Drivers shown:

```
• EMI load is moderately high
• Low financial buffer
• Multiple short-term debts
```

---

# 13. Health Timeline

Timeline is generated from snapshots.

Query:

```
SELECT created_at, score
FROM financial_snapshots
ORDER BY created_at;
```

This produces a trend graph:

```
Financial Health Over Time
```

---

# 14. System Architecture

Final architecture layers:

```
Identity Layer
    users

Debt Portfolio Layer
    debts
    payments_history

Financial State Layer
    financial_snapshots

Decision Intelligence Layer
    classification engine
    stress simulation

Insight Layer
    financial health score
```

---

# 15. Key Design Rules

### Store inputs, not conclusions

Never store:

```
health_score
stress_level
classification_result
```

These must always be recomputed.

---

### Simulations must remain isolated

Hypothetical loans must never modify real debt records.

---

### Financial snapshots are authoritative state

All engines depend on snapshots as their input.

---

# 16. Summary

This architecture establishes a **Financial State Engine** at the center of Debtease.

Benefits:

* unified onboarding and logged-in logic
* deterministic financial analysis
* future-proof system design
* support for timeline insights
* clean separation between state and intelligence

The system can now evolve from a **loan analysis tool** into a **full debt management platform** without major architectural changes.
