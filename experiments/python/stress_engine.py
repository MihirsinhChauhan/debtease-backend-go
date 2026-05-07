from __future__ import annotations

from dataclasses import dataclass
from typing import Dict, List, Literal, Sequence, Tuple

from loan_portfolio_engine import EMIRecord, Loan


StressLevel = Literal["RESILIENT", "SENSITIVE", "FRAGILE"]
SurvivabilityBand = Literal["IMMEDIATE_STRESS", "SHORT", "MODERATE", "STABLE"]


@dataclass
class DebtSnapshot:
    loan_id: str
    emi: float
    remaining_months: int
    instrument_type: str
    secured: bool


@dataclass
class StressOutcome:
    stress_level: StressLevel
    flexibility_after_shock: float
    survivability_band: SurvivabilityBand
    failed_debt_categories: List[str]
    key_drivers: List[str]


# ----------------------------
# Helpers
# ----------------------------

def _build_snapshots(
    schedule: Dict[int, List[EMIRecord]],
    loans: List[Loan],
    month: int,
) -> List[DebtSnapshot]:
    """Represent each debt at a given month in the minimal form the stress engine needs."""
    last_month: Dict[str, int] = {}
    for m, records in schedule.items():
        for r in records:
            if r.loan_id not in last_month or m > last_month[r.loan_id]:
                last_month[r.loan_id] = m

    loan_index: Dict[str, Loan] = {loan.loan_id: loan for loan in loans}
    snapshots: List[DebtSnapshot] = []
    for r in schedule.get(month, []):
        loan = loan_index.get(r.loan_id)
        if loan is None:
            continue
        remaining = max(0, last_month.get(r.loan_id, month) - month + 1)
        snapshots.append(
            DebtSnapshot(
                loan_id=r.loan_id,
                emi=r.emi,
                remaining_months=remaining,
                instrument_type=getattr(loan, "instrument_type", "other"),
                secured=bool(getattr(loan, "secured", False)),
            )
        )
    return snapshots


def _flexibility_buffer(
    income: float, fixed_expenses: float, total_emi: float
) -> Tuple[float, float]:
    """
    Returns (buffer_amount, months_of_buffer).
    Months_of_buffer approximates how many months EMIs can be sustained from surplus.
    """
    disposable = income - fixed_expenses
    buffer = disposable - total_emi
    if total_emi <= 0:
        return buffer, float("inf")
    months = buffer / total_emi
    return buffer, months


def _classify_survivability(months_of_buffer: float) -> SurvivabilityBand:
    if months_of_buffer < 1:
        return "IMMEDIATE_STRESS"
    if months_of_buffer < 3:
        return "SHORT"
    if months_of_buffer < 6:
        return "MODERATE"
    return "STABLE"


def _classify_stress_level(buffer: float, months_of_buffer: float) -> StressLevel:
    if buffer < 0:
        return "FRAGILE"
    if months_of_buffer < 3:
        return "SENSITIVE"
    return "RESILIENT"


def _category_for_debt(d: DebtSnapshot) -> str:
    if d.instrument_type == "bnpl":
        return "BNPL_OR_ZERO_EMI"
    if not d.secured and d.remaining_months <= 12:
        return "SHORT_TERM_UNSECURED"
    if not d.secured:
        return "LONG_TERM_UNSECURED"
    return "SECURED"


def _apply_interest_rate_shock(debts: List[DebtSnapshot]) -> List[float]:
    """
    Returns per-debt EMI' after interest rate shock.

    Simplified v1 policy:
    - BNPL / 0% EMI (instrument_type='bnpl') -> no rate shock (β=0)
    - Secured loans (e.g. mortgage) -> treated as fixed-rate here (β=0)
    - Other unsecured loans -> mild increase (β=0.1)
    """
    shocked_emis: List[float] = []
    for d in debts:
        if d.instrument_type == "bnpl":
            beta = 0.0
        elif d.secured:
            beta = 0.0
        else:
            beta = 0.1
        shocked_emis.append(d.emi * (1.0 + beta))
    return shocked_emis


def _failure_order(
    buffer_after_shock: float,
    shocked_emis: List[float],
    debts: List[DebtSnapshot],
) -> List[str]:
    """
    When buffer is negative, consume deficit by dropping EMIs in fixed priority order.
    Returns the ordered list of distinct categories that would fail.
    """
    if buffer_after_shock >= 0:
        return []

    deficit = -buffer_after_shock

    # Priority: BNPL / 0% EMI → short-term unsecured → long-term unsecured → secured
    ordered_debts = sorted(
        debts,
        key=lambda d: {
            "BNPL_OR_ZERO_EMI": 0,
            "SHORT_TERM_UNSECURED": 1,
            "LONG_TERM_UNSECURED": 2,
            "SECURED": 3,
        }[_category_for_debt(d)],
    )

    failed_categories: List[str] = []
    for idx, d in enumerate(ordered_debts):
        emi_prime = shocked_emis[idx]
        deficit -= emi_prime
        category = _category_for_debt(d)
        if category not in failed_categories:
            failed_categories.append(category)
        if deficit <= 0:
            break

    return failed_categories


def simulate_stress(
    income: float,
    fixed_expenses: float,
    schedule: Dict[int, List[EMIRecord]],
    loans: List[Loan],
    month: int = 1,
    alphas: Sequence[float] = (0.9, 0.8),
) -> StressOutcome:
    """
    Run deterministic stress scenarios over a portfolio snapshot.

    - Income' = income * α for α in {0.9, 0.8}
    - Interest rate shock is applied selectively to unsecured loans.
    - We then compute flexibility buffer, survivability band, failure order,
      and key drivers based on the worst-case (most fragile) scenario.
    """
    debts = _build_snapshots(schedule, loans, month)
    if not debts:
        # No debts -> resilient by definition for this domain.
        return StressOutcome(
            stress_level="RESILIENT",
            flexibility_after_shock=income - fixed_expenses,
            survivability_band="STABLE",
            failed_debt_categories=[],
            key_drivers=[],
        )

    shocked_emis = _apply_interest_rate_shock(debts)
    total_shocked_emi = sum(shocked_emis)

    worst_level: StressLevel = "RESILIENT"
    worst_buffer: float = 0.0
    worst_months: float = 0.0

    def _severity(level: StressLevel) -> int:
        if level == "FRAGILE":
            return 2
        if level == "SENSITIVE":
            return 1
        return 0

    for alpha in alphas:
        stressed_income = income * alpha
        buffer, months = _flexibility_buffer(
            stressed_income, fixed_expenses, total_shocked_emi
        )
        level = _classify_stress_level(buffer, months)
        if _severity(level) > _severity(worst_level):
            worst_level = level
            worst_buffer = buffer
            worst_months = months

    survivability = _classify_survivability(worst_months)
    failed_categories = _failure_order(worst_buffer, shocked_emis, debts)

    # Key drivers (max 3)
    drivers: List[str] = []
    if worst_buffer < 0:
        drivers.append("Low flexibility buffer")
    elif worst_months < 3:
        drivers.append("Thin flexibility buffer")

    unsecured_emi = sum(
        shocked_emis[idx]
        for idx, d in enumerate(debts)
        if not d.secured and _category_for_debt(d) != "BNPL_OR_ZERO_EMI"
    )
    if total_shocked_emi > 0 and unsecured_emi / total_shocked_emi > 0.5:
        drivers.append("Unsecured exposure")

    short_term_unsecured_emi = sum(
        shocked_emis[idx]
        for idx, d in enumerate(debts)
        if _category_for_debt(d) == "SHORT_TERM_UNSECURED"
    )
    if total_shocked_emi > 0 and short_term_unsecured_emi / total_shocked_emi > 0.3:
        drivers.append("High short-term rigidity")

    drivers = drivers[:3]

    return StressOutcome(
        stress_level=worst_level,
        flexibility_after_shock=worst_buffer,
        survivability_band=survivability,
        failed_debt_categories=failed_categories,
        key_drivers=drivers,
    )


__all__ = [
    "StressLevel",
    "SurvivabilityBand",
    "DebtSnapshot",
    "StressOutcome",
    "simulate_stress",
]

