from decimal import Decimal
from typing import List

from classification_engine import (
    ACTIVELY_REDUCE,
    AVOID_ADDING,
    FLAG_HIGH_CASHFLOW_RIGIDITY,
    FLAG_SHORT_TERM_DEBT_STACKING,
    ClassificationInput,
    DebtEntry,
    FinancialContext,
    SAFE_TO_HOLD,
    classify,
    INSTRUMENT_BNPL,
    INSTRUMENT_CREDIT_CARD,
    INSTRUMENT_MORTGAGE,
)


def _has_flag(flags: List[str], flag: str) -> bool:
    return flag in flags


def _make_input(
    income: str,
    fixed_obligations: str,
    debts: List[DebtEntry],
) -> ClassificationInput:
    total_emi = sum(d.emi for d in debts)
    ctx = FinancialContext(
        monthly_net_income=Decimal(income),
        fixed_obligations=Decimal(fixed_obligations),
        existing_emis=total_emi,
        short_term_obligations_count=0,
    )
    return ClassificationInput(debts=debts, context=ctx)


def _debt_entry(
    loan_id: str,
    emi: str,
    remaining_months: int,
    secured: bool,
    instrument_type: str,
) -> DebtEntry:
    return DebtEntry(
        id=loan_id,
        emi=Decimal(emi),
        remaining_months=remaining_months,
        secured=secured,
        instrument_type=instrument_type,
    )


def test_clean_borrower_safe_to_hold_no_flags():
    # Income 100k, EMI 20k -> ratio 0.2. One mortgage (secured), no BNPL.
    debts = [
        _debt_entry("hl", "20000", 120, True, INSTRUMENT_MORTGAGE),
    ]
    input_data = _make_input("100000", "5000", debts)
    result = classify(input_data)

    assert result.treatment == SAFE_TO_HOLD
    assert result.flags == []


def test_moderate_stress_flags_high_cashflow_rigidity():
    # Income 100k, EMI 40k -> ratio 0.4. Buffer at threshold.
    debts = [
        _debt_entry("d1", "25000", 24, False, "personal_loan"),
        _debt_entry("d2", "15000", 36, False, INSTRUMENT_CREDIT_CARD),
    ]
    input_data = _make_input("100000", "0", debts)
    result = classify(input_data)

    assert result.treatment in (ACTIVELY_REDUCE, "NEEDS_MONITORING")
    assert _has_flag(result.flags, FLAG_HIGH_CASHFLOW_RIGIDITY)


def test_debt_loop_pattern_avoid_adding():
    # 5 BNPL -> fragmentation 5 + 5*0.5 = 7.5 >= 4. EMI ratio kept moderate so we get AVOID_ADDING from stacking.
    debts = [
        _debt_entry("b1", "2000", 2, False, INSTRUMENT_BNPL),
        _debt_entry("b2", "1500", 2, False, INSTRUMENT_BNPL),
        _debt_entry("b3", "1000", 3, False, INSTRUMENT_BNPL),
        _debt_entry("b4", "2500", 2, False, INSTRUMENT_BNPL),
        _debt_entry("b5", "500", 1, False, INSTRUMENT_BNPL),
    ]
    input_data = _make_input("80000", "0", debts)
    result = classify(input_data)

    assert result.treatment == AVOID_ADDING
    assert _has_flag(result.flags, FLAG_SHORT_TERM_DEBT_STACKING)


def test_dangerous_structure_actively_reduce():
    # High unsecured (100%), low buffer: income 60k, fixed 5k, EMI 40k -> surplus 15k, buffer 15/40 < 1.5.
    debts = [
        _debt_entry("d1", "25000", 48, False, "personal_loan"),
        _debt_entry("d2", "15000", 24, False, INSTRUMENT_CREDIT_CARD),
    ]
    input_data = _make_input("60000", "5000", debts)
    result = classify(input_data)

    assert result.treatment == ACTIVELY_REDUCE
    assert len(result.flags) >= 2

