#!/usr/bin/env python3
"""Independent FIDES 2022 reference calculation for the fitcalc validation (M1.10).

The models of the probe circuit's components are implemented here directly
from the FIDES Guide 2022 Edition A (July 2023), without the Go fides library.
Page numbers refer to that guide. The stresses (power, voltage, junction
temperature rise) are the ones fitcalc derives from ngspice for
fitcalc's testdata/probe.net, probe.bom.csv and parts.db.csv, so that only the
reliability model is compared.

Usage:
  fides_ref.py                 writes the reference FIT values (reference.csv)
  fides_ref.py fitcalc.csv     compares them with a fitcalc CSV report of the
                               probe circuit
"""

import csv
import math
import os
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
MISSION = os.path.join(HERE, "..", "mission.csv")

K = 11604            # 1/k_B in K/eV, as rounded in the guide (p. 41)
PM_ACTIVE = 1.7      # ΠPM default: ICs, discrete semiconductors, LEDs, optocouplers (p. 36)
PM_PASSIVE = 1.6     # ΠPM default: other components (p. 36)
PROCESS = 4.0        # ΠProcess default (p. 37)
RUGGEDIZING = 1.7    # ΠRuggedizing default (p. 115)

# Pollution levels (p. 180). Saline pollution has no moderate level in the
# guide; 1.5 is used for it, as in the Go library.
SALINE = {"low": 1, "moderate": 1.5, "high": 2}
ENVIRONMENT = {"low": 1, "moderate": 1.5, "high": 2}
ZONE = {"low": 1, "moderate": 2, "high": 4}


def mission(path=MISSION):
    """The mission phases of a mission profile file, by default
    testdata/mission.csv (the cmd/fides sample)."""
    phases = []
    with open(path) as f:
        rows = csv.reader(f)
        head = [h.strip() for h in next(rows)]
        for r in rows:
            p = dict(zip(head, (x.strip() for x in r)))
            phases.append({
                "t": float(p["duration"]), "on": p["on"] == "on",
                "tamb": float(p["tamb"]), "rh": float(p["rh"]),
                "dt": float(p["tdelta"]), "n": float(p["ncycles"]),
                "theta": float(p["tcycle"]), "tmax": float(p["tmax"]),
                "grms": float(p["grms"]), "app": float(p["pi_app"]),
                "sal": SALINE[p["saline_pollution"]],
                "envir": ENVIRONMENT[p["env_pollution"]],
                "zone": ZONE[p["app_pollution"]],
                "prot": 0 if p["ip"] in ("sealed", "hermetic") else 1,
            })
    return phases


def cs(eos, mos, tos):
    """Csensitivity (p. 110). The datasheet tables print it rounded to two
    decimals."""
    return 0.725 * eos + 0.225 * mos + 0.05 * tos


def arrhenius(ea, temp):
    """Acceleration relative to 20 °C (p. 41)."""
    return math.exp(K * ea * (1 / 293 - 1 / (temp + 273)))


def pi_tcy_solder(ph):
    """Thermal cycling of solder joints (pp. 44, 129, 137, 148)."""
    return (12 * ph["n"] / ph["t"] * (min(ph["theta"], 2) / 2) ** (1 / 3)
            * (ph["dt"] / 20) ** 1.9 * math.exp(1414 * (1 / 313 - 1 / (ph["tmax"] + 273))))


def pi_tcy_case(ph):
    """Thermal cycling of the case of a semiconductor (pp. 129, 137)."""
    return (12 * ph["n"] / ph["t"] * (ph["dt"] / 20) ** 4
            * math.exp(1414 * (1 / 313 - 1 / (ph["tmax"] + 273))))


def pi_mech(ph):
    """Vibration (p. 49)."""
    return (ph["grms"] / 0.5) ** 1.5


def pi_rh(ea, ph):
    """Humidity (Peck, p. 47)."""
    return (ph["rh"] / 70) ** 4.4 * arrhenius(ea, ph["tamb"])


def fit(phases, physical, csens, placement, pm):
    """λ = λPhysical × ΠPM × ΠProcess, λPhysical weighted by phase duration,
    with the induced factor per phase (pp. 30-31, 110)."""
    total = sum(ph["t"] for ph in phases)
    lp = 0
    for ph in phases:
        induced = (placement * ph["app"] * RUGGEDIZING) ** (0.511 * math.log(csens))
        lp += ph["t"] / total * physical(ph) * induced
    return lp * pm * PROCESS


def resistor_thick_smd(phases, p, prated):
    """Fixed resistor, SMD, thick film (pp. 146-148)."""
    l0, a, g_th, g_tcy, g_mech, g_rh = 0.01, 70, 0.01, 0.97, 0.01, 0.01

    def physical(ph):
        th = g_th * arrhenius(0.15, ph["tamb"] + a * p / prated) if ph["on"] else 0
        rh = 0 if ph["on"] else g_rh * pi_rh(0.9, ph)
        return l0 * (th + g_tcy * pi_tcy_solder(ph) + g_mech * pi_mech(ph) + rh)

    return fit(phases, physical, cs(4, 3, 5), 1.0, PM_PASSIVE)


def ceramic_type2_cat1(phases, v, vrated):
    """Ceramic capacitor, type II (X7R), category 1 (pp. 151-153)."""
    l0, ea, sref, g_th, g_tcy, g_mech = 0.08, 0.1, 0.3, 0.70, 0.28, 0.02

    def physical(ph):
        th = g_th * (v / vrated / sref) ** 3 * arrhenius(ea, ph["tamb"]) if ph["on"] else 0
        return l0 * (th + g_tcy * pi_tcy_solder(ph) + g_mech * pi_mech(ph))

    return fit(phases, physical, cs(7, 6, 1), 1.0, PM_PASSIVE)


def inductor_power(phases):
    """High-current (power) wirewound inductor, with the table's typical
    temperature rise (pp. 161-162)."""
    l0, ea, g_th, g_tcy, g_mech, dt = 0.05, 0.15, 0.09, 0.79, 0.12, 30

    def physical(ph):
        th = g_th * arrhenius(ea, ph["tamb"] + dt) if ph["on"] else 0
        return l0 * (th + g_tcy * pi_tcy_solder(ph) + g_mech * pi_mech(ph))

    return fit(phases, physical, cs(7, 6, 3), 1.0, PM_PASSIVE)


# Package basic failure rates λ0RH, λ0TCy_case, λ0TCy_solder, λ0Mech of
# "SMD, small signal, L-lead, plastic": SOT23, SOD123, … (p. 135)
SMD_SMALL_SIGNAL = (0.002467, 0.000048, 0.000242, 0.000005)

# Chip basic failure rates λ0TH (p. 136)
CHIP = {"signal diode": 0.0044, "bjt": 0.0138, "mos": 0.0145, "jfet": 0.0143}


def discrete(phases, chip, rise, vr=None, vrated=None):
    """Discrete semiconductor in an SMD small signal L-lead plastic package
    (pp. 133-137). rise is the junction temperature rise RJA·P; vr and vrated
    are the reverse voltage and its rating, for signal diodes."""
    l_rh, l_case, l_solder, l_mech = SMD_SMALL_SIGNAL
    pi_el = 1
    if vr is not None:
        ratio = vr / vrated
        pi_el = ratio ** 2.4 if ratio > 0.3 else 0.056

    def physical(ph):
        th = pi_el * arrhenius(0.7, ph["tamb"] + rise) if ph["on"] else 0
        rh = 0 if ph["on"] else l_rh * pi_rh(0.9, ph)
        return (CHIP[chip] * th + l_case * pi_tcy_case(ph) + l_solder * pi_tcy_solder(ph)
                + rh + l_mech * pi_mech(ph))

    return fit(phases, physical, cs(8, 2, 1), 1.0, PM_ACTIVE)


def ic_so(phases, np, l0th, rise, placement):
    """Integrated circuit in an SO (SOIC) package with up to 18 pins
    (pp. 123-129)."""
    def rate(a, b):
        return math.exp(-a) * np ** b

    l_rh, l_case = rate(11.45, 1.95), rate(16.80, 2.94)
    l_solder, l_mech = rate(8.29, 0.92), rate(12.90, 0.92)

    def physical(ph):
        th = arrhenius(0.7, ph["tamb"] + rise) if ph["on"] else 0
        rh = 0 if ph["on"] else l_rh * pi_rh(0.9, ph)
        return (l0th * th + l_case * pi_tcy_case(ph) + l_solder * pi_tcy_solder(ph)
                + rh + l_mech * pi_mech(ph))

    return fit(phases, physical, cs(10, 2, 1), placement, PM_ACTIVE)


def connector_pcb(phases, contacts, mount, dt):
    """Connector for printed circuits, fewer than one connection cycle a year
    (pp. 177-180). The placement factor of connectors is 1."""
    l0 = 0.1 * mount * contacts ** 0.5 * 0.2

    def physical(ph):
        th = 0.58 * arrhenius(0.1, ph["tamb"] + dt) if ph["on"] else 0
        chem = 0.20 * ph["sal"] * ph["envir"] * ph["zone"] * ph["prot"]
        return l0 * (th + 0.04 * pi_tcy_solder(ph) + 0.05 * pi_mech(ph)
                     + 0.13 * pi_rh(0.8, ph) + chem)

    return fit(phases, physical, cs(1, 10, 3), 1.0, PM_PASSIVE)


# The probe circuit's components with the stresses fitcalc derives.
# Placement: digital non-interface (1.0), except X1, tagged analog (1.3).
CASES = {
    "R1": lambda m: resistor_thick_smd(m, 0.0119008, 0.1),
    "R3": lambda m: resistor_thick_smd(m, 0.127288, 0.25),
    "C1": lambda m: ceramic_type2_cat1(m, 1.09091, 50),
    "L1": lambda m: inductor_power(m),
    "D1": lambda m: discrete(m, "signal diode", 2.72913, vr=0, vrated=100),
    "D2": lambda m: discrete(m, "signal diode", 4.85684e-08, vr=12, vrated=100),
    "Q1": lambda m: discrete(m, "bjt", 0.36501),
    "M1": lambda m: discrete(m, "mos", 7.23488),
    "J1": lambda m: discrete(m, "jfet", 1.772),
    "X1": lambda m: ic_so(m, 8, 0.086, 0, 1.3),
    "J9": lambda m: connector_pcb(m, 2, 6, 0),
}


def main():
    phases = mission()
    w = csv.writer(sys.stdout)

    if len(sys.argv) < 2:
        w.writerow(["ref", "fit"])
        for ref, case in CASES.items():
            w.writerow([ref, "%.9g" % case(phases)])
        return

    with open(sys.argv[1]) as f:
        got = {r["ref"]: float(r["fit"]) for r in csv.DictReader(f) if r["fit"]}
    w.writerow(["ref", "guide", "fitcalc", "fitcalc/guide"])
    for ref, case in CASES.items():
        ref_fit = case(phases)
        w.writerow([ref, "%.6g" % ref_fit, "%.6g" % got[ref], "%.6f" % (got[ref] / ref_fit)])


if __name__ == "__main__":
    main()
