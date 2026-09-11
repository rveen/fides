#!/usr/bin/env python3
"""FIDES 2022 reference FIT values for the component families of the fides
library (fitcalc M1.10).

Each row of cases.csv names the guide model and table row that applies
("model"), the placement factor and the mission profile, and gives the
inputs. This script computes the FIT from the FIDES Guide 2022 Edition A
(July 2023), page numbers as noted, without the Go fides library. The Go test
guide_cases_test.go runs the same rows through the library, which has to
choose the same model from the class, tags and package.

Usage: fides_cases.py > reference-cases.csv
"""

import csv
import math
import os
import sys

from fides_ref import (K, PM_ACTIVE, PM_PASSIVE, PROCESS, RUGGEDIZING, arrhenius, cs,
                       mission, pi_mech, pi_rh, pi_tcy_case, pi_tcy_solder)

HERE = os.path.dirname(os.path.abspath(__file__))
TESTDATA = os.path.join(HERE, "..")


def fit(phases, physical, csens, placement, pm, pw=1.0):
    """λ = λPhysical × ΠPW × ΠPM × ΠProcess (pp. 30-31, 110, 138)."""
    total = sum(ph["t"] for ph in phases)
    lp = 0
    for ph in phases:
        induced = (placement * ph["app"] * RUGGEDIZING) ** (0.511 * math.log(csens))
        lp += ph["t"] / total * physical(ph) * induced
    return lp * pw * pm * PROCESS


# Resistors (pp. 146-148): λ0, A (°C), γTH-EL, γTCy, γMech, γRH, (EOS, MOS, TOS)
RESISTORS = {
    "melf": (0.1, 85, 0.04, 0.89, 0.01, 0.06, (4, 2, 4)),
    "thick_power": (0.4, 130, 0.04, 0.89, 0.01, 0.06, (2, 4, 1)),
    "ww": (0.3, 30, 0.02, 0.96, 0.01, 0.01, (2, 1, 3)),
    "ww_power": (0.4, 130, 0.01, 0.97, 0.01, 0.01, (2, 4, 1)),
    "pot": (0.3, 65, 0.42, 0.35, 0.22, 0.01, (1, 5, 2)),
    "thick": (0.01, 70, 0.01, 0.97, 0.01, 0.01, (4, 3, 5)),
    "network": (0.01, 70, 0.01, 0.97, 0.01, 0.01, (3, 5, 3)),
    "thin_smd_lo": (0.18, 85, 0.14, 0.53, 0.07, 0.26, (5, 5, 4)),
    "thin_smd_mid": (0.21, 85, 0.10, 0.54, 0.06, 0.30, (5, 5, 4)),
    "thin_smd_hi": (0.25, 85, 0.07, 0.55, 0.05, 0.33, (5, 5, 4)),
    "thin_tht_lo": (0.14, 85, 0.18, 0.43, 0.08, 0.31, (5, 5, 4)),
    "thin_tht_mid": (0.18, 85, 0.12, 0.44, 0.07, 0.37, (5, 5, 4)),
    "thin_tht_hi": (0.21, 85, 0.08, 0.45, 0.06, 0.41, (5, 5, 4)),
}


def resistor(row, r, phases):
    l0, a, g_th, g_tcy, g_mech, g_rh, sens = RESISTORS[row]
    if r["n"] > 1:
        l0 *= math.sqrt(r["n"])  # networks: 0.01 × √NR
    rise = a * r["p"] / r["pmax"]

    def physical(ph):
        th = g_th * arrhenius(0.15, ph["tamb"] + rise) if ph["on"] else 0
        rh = 0 if ph["on"] else g_rh * pi_rh(0.9, ph)
        return l0 * (th + g_tcy * pi_tcy_solder(ph) + g_mech * pi_mech(ph) + rh)

    return fit(phases, physical, cs(*sens), r["placement"], PM_PASSIVE)


# Capacitors (pp. 151-157): λ0, Ea, Sreference, γTH-EL, γTCy, γMech, (EOS, MOS, TOS)
CAPACITORS = {
    "ceramic/type1_1": (0.03, 0.1, 0.3, 0.70, 0.28, 0.02, (7, 5, 2)),
    "ceramic/type1_2": (0.05, 0.1, 0.3, 0.70, 0.28, 0.02, (7, 5, 2)),
    "ceramic/type1_3": (0.40, 0.1, 0.3, 0.69, 0.26, 0.05, (7, 5, 2)),
    "ceramic/type2_1": (0.08, 0.1, 0.3, 0.70, 0.28, 0.02, (7, 6, 1)),
    "ceramic/type2_2": (0.15, 0.1, 0.3, 0.70, 0.28, 0.02, (7, 6, 1)),
    "ceramic/type2_3": (1.20, 0.1, 0.3, 0.44, 0.51, 0.05, (7, 6, 1)),
    "ceramic/flex_x5r_1": (0.08, 0.1, 0.3, 0.70, 0.28, 0.02, (7, 4, 2)),
    "ceramic/flex_x7r_1": (0.08, 0.1, 0.3, 0.70, 0.28, 0.02, (7, 4, 1)),
    "ceramic/flex_x5r_23": (0.15, 0.1, 0.3, 0.70, 0.28, 0.02, (7, 4, 2)),
    "ceramic/flex_x7r_23": (0.15, 0.1, 0.3, 0.70, 0.28, 0.02, (7, 4, 1)),
    "alu/wet": (0.21, 0.40, 0.5, 0.85, 0.14, 0.01, (7, 7, 1)),
    "alu/solid": (0.4, 0.40, 0.55, 0.85, 0.14, 0.01, (7, 7, 1)),
    "tant/wet_elastomer": (0.77, 0.15, 0.6, 0.87, 0.01, 0.12, (8, 7, 1)),
    "tant/wet_silver_glass": (0.33, 0.15, 0.6, 0.81, 0.01, 0.18, (8, 7, 1)),
    "tant/wet_tantalum_case": (0.05, 0.15, 0.6, 0.88, 0.04, 0.08, (8, 7, 1)),
    "tant/solid_bead": (1.09, 0.15, 0.4, 0.86, 0.12, 0.02, (8, 7, 1)),
    "tant/solid_smd": (0.54, 0.15, 0.4, 0.84, 0.14, 0.02, (8, 7, 1)),
    "tant/solid_axial": (0.25, 0.15, 0.4, 0.94, 0.04, 0.02, (8, 7, 1)),
}


def capacitor(model, r, phases):
    l0, ea, sref, g_th, g_tcy, g_mech, sens = CAPACITORS[model]

    def physical(ph):
        th = g_th * (r["v"] / r["vmax"] / sref) ** 3 * arrhenius(ea, ph["tamb"]) if ph["on"] else 0
        return l0 * (th + g_tcy * pi_tcy_solder(ph) + g_mech * pi_mech(ph))

    return fit(phases, physical, cs(*sens), r["placement"], PM_PASSIVE)


# Magnetic components (pp. 161-162): λ0, Ea, γTH-EL, γTCy, γMech, ΔT, (EOS, MOS, TOS)
INDUCTORS = {
    "ww_low": (0.025, 0.15, 0.01, 0.73, 0.26, 10, (5, 4, 4)),
    "ww_power": (0.05, 0.15, 0.09, 0.79, 0.12, 30, (7, 6, 3)),
    "multilayer": (0.05, 0.15, 0.71, 0.28, 0.01, 10, (4, 6, 1)),
    "trafo_low": (0.125, 0.15, 0.01, 0.73, 0.26, 10, (6, 5, 3)),
    "trafo_power": (0.25, 0.15, 0.15, 0.69, 0.16, 30, (6, 7, 4)),
}


def inductor(row, r, phases):
    l0, ea, g_th, g_tcy, g_mech, dt, sens = INDUCTORS[row]

    def physical(ph):
        th = g_th * arrhenius(ea, ph["tamb"] + dt) if ph["on"] else 0
        return l0 * (th + g_tcy * pi_tcy_solder(ph) + g_mech * pi_mech(ph))

    return fit(phases, physical, cs(*sens), r["placement"], PM_PASSIVE)


# Discrete semiconductor packages (pp. 135-136): λ0RH, λ0TCyCase, λ0TCySolder, λ0Mech
PACKAGES = {
    "tht_signal": (0.005713, 0.000298, 0.001491, 0.000030),
    "smd_signal": (0.002467, 0.000048, 0.000242, 0.000005),
    "smd_medium": (0.003009, 0.000747, 0.003734, 0.000075),
    "tht_power": (0.039622, 0.001553, 0.007764, 0.000155),
    "smd_clead": (0.010994, 0.000590, 0.002949, 0.000059),
    "smd_power": (0.027131, 0.001296, 0.006479, 0.000130),
    "screw_power": (0.487813, 0.013371, 0.066853, 0.001337),
    "smd_glass": (0, 0.004788, 0.023938, 0.000479),
    "tht_metal": (0, 0.004247, 0.021234, 0.000425),
    "smd_flat": (0.033236, 0.003246, 0.016230, 0.000325),
}

# Discrete chip λ0TH (pp. 136, 139)
CHIPS = {
    "signal": 0.0044, "rectifier": 0.0100, "rectifier_power": 0.1574,
    "zener": 0.0080, "zener_power": 0.0954, "tvs": 0.0210, "tvs_power": 1.4980,
    "thyristor": 0.1976, "bjt": 0.0138, "bjt_power": 0.0478, "mos": 0.0145,
    "jfet": 0.0143, "power_mos": 0.56, "power_igbt": 0.56,
}


def discrete(chip, group, r, phases, power=False):
    """Discrete semiconductor (pp. 133-137); power: silicon MOS over 5 W and
    IGBTs (pp. 138-140)."""
    l_rh, l_case, l_solder, l_mech = PACKAGES[group]
    l0th = CHIPS[chip] * math.sqrt(max(r["n"], 1))
    pi_el = 1
    if chip == "signal":
        ratio = r["v"] / r["vmax"]
        pi_el = ratio ** 2.4 if ratio > 0.3 else 0.056

    def physical(ph):
        th = 0
        if ph["on"]:
            tj = ph["tamb"] + r["t"]
            if power:
                tj = min(tj, 175)
                th = math.exp(K * 0.7 * (1 / (60 + 273) - 1 / (tj + 273)))
            else:
                th = pi_el * arrhenius(0.7, tj)
        rh = 0 if ph["on"] else l_rh * pi_rh(0.9, ph)
        return (l0th * th + l_case * pi_tcy_case(ph) + l_solder * pi_tcy_solder(ph)
                + rh + l_mech * pi_mech(ph))

    return fit(phases, physical, cs(8, 2, 1), r["placement"], PM_ACTIVE, 10 if power else 1)


# LED packages (p. 142): λ0RH, λ0TCyCase, λ0TCySolder, λ0Mech
LED_PACKAGES = {
    "low_std": (0.0034, 0.0104, 0.0520, 0.0052),
    "low_round": (0.0034, 0.0104, 0.1560, 0.0624),
    "low_lga_plastic": (0.0034, 0.0104, 0.2080, 0.0832),
    "low_lga_ceramic": (0.0034, 0.0104, 0.3640, 0.1820),
    "low_other_plastic": (0.0034, 0.0104, 0.1560, 0.0624),
    "low_other_ceramic": (0.0034, 0.0104, 0.3640, 0.1820),
    "high_plastic": (0.0031, 0.0042, 0.0420, 0.0064),
    "high_ceramic": (0.0031, 0.0042, 0.1470, 0.0735),
}


def led(colour, pkg, r, phases):
    """Light-emitting diode (pp. 141-143)."""
    l_rh, l_case, l_solder, l_mech = LED_PACKAGES[pkg]
    l0th = (0.05 if colour == "white" else 0.01) * math.sqrt(max(r["n"], 1))

    def physical(ph):
        th = arrhenius(0.4, ph["tamb"] + r["t"]) if ph["on"] else 0
        rh = 0 if ph["on"] else l_rh * pi_rh(0.9, ph)
        return (l0th * th + l_case * pi_tcy_case(ph) + l_solder * pi_tcy_solder(ph)
                + rh + l_mech * pi_mech(ph))

    return fit(phases, physical, cs(7, 2, 3), r["placement"], PM_ACTIVE)


# Integrated circuit packages (pp. 125-127): λ0RH (a, b) or None (hermetic),
# λ0TCyCase (a, b), and the solder joint and mechanical a for pin ranges up
# to the given number of pins; b = 0.92 for both
IC_PACKAGES = {
    "PDIP": ((6.27, 0.69), (10.23, 0.95), [(68, 8.29, 12.90)]),
    "CERDIP": (None, (12.68, 2.27), [(20, 8.29, 11.51), (48, 7.96, 11.18)]),
    "PQFP": ((10.94, 1.57), (13.72, 1.62), [(240, 8.29, 12.21), (304, 7.96, 11.87)]),
    "LQFP": ((6.62, 0.52), (13.05, 1.30), [(120, 8.29, 12.90), (208, 7.20, 11.80)]),
    "POWERQFP": ((14.17, 2.41), (11.80, 1.36), [(240, 8.29, 12.21), (304, 7.96, 11.87)]),
    "CERPACK": (None, (8.14, 1.01), [(56, 8.29, 11.51)]),
    "CQFP": (None, (8.14, 1.01), [(132, 8.29, 11.51), (256, 6.68, 9.90)]),
    "PLCC": ((10.50, 1.92), (19.45, 3.29), [(52, 8.29, 12.43), (84, 7.20, 11.29)]),
    "JLCC": (None, (8.07, 0.93), [(32, 8.29, 11.51), (44, 7.95, 11.17), (52, 7.19, 10.41),
                                  (68, 6.21, 9.43), (84, 5.58, 8.80)]),
    "CLCC": (None, (8.07, 0.93), [(10, 7.19, 10.41), (20, 5.89, 9.11), (32, 5.58, 8.80),
                                  (52, 5.07, 8.29), (84, 4.36, 7.58)]),
    "SOJ": ((4.33, 0.83), (8.76, 1.49), [(44, 8.29, 12.90)]),
    "SO": ((11.45, 1.95), (16.80, 2.94), [(18, 8.29, 12.90), (32, 7.96, 12.56)]),
    "TSOP": ((6.87, 1.10), (9.60, 0.83), [(16, 8.29, 12.90), (32, 7.20, 11.80), (44, 6.68, 11.29),
                                          (56, 6.17, 10.77)]),
    "SSOP": ((17.70, 3.35), (20.88, 3.38), [(64, 7.96, 12.56)]),
    "TSSOP": ((11.25, 1.57), (14.93, 1.87), [(28, 8.29, 12.90), (48, 7.96, 12.56), (56, 7.20, 11.80),
                                             (64, 6.17, 10.77)]),
    "QFN": ((8.84, 0.77), (12.03, 0.94), [(24, 6.68, 9.90), (56, 6.17, 9.38), (80, 5.95, 9.64)]),
    "QFN04": ((6.22, 0.78), (9.65, 0.91), [(40, 6.17, 9.38), (80, 5.95, 9.17)]),
    "LGA": ((5.3, 0.84), (9.02, 1.02), [(20, 6.68, 9.90), (10 ** 6, 6.17, 9.38)]),
    "PBGA": ((6.85, 0.80), (10.29, 0.86), [(352, 7.20, 10.87), (432, 6.68, 10.37), (729, 6.17, 9.85)]),
    "PBGABT": ((4.77, 0.40), (10.46, 0.76), [(484, 6.68, 10.37), (1156, 6.17, 9.85)]),
    "POWERBGA": ((5.59, 0.61), (11.01, 0.79), [(352, 7.20, 10.87), (956, 6.68, 10.37)]),
    "FCBGA": ((8.39, 0.79), (9.82, 0.52), [(671, 7.20, 10.87), (1704, 6.68, 10.37)]),
    "CBGA": ((4.41, 0.63), (10.60, 1.12), [(560, 5.12, 8.80), (1156, 4.36, 7.35)]),
    "DBGA": ((4.41, 0.63), (10.60, 1.12), [(1156, 6.68, 10.37)]),
    "CCGA": ((4.41, 0.63), (10.60, 1.12), [(1156, 5.95, 9.64)]),
    "CPGA": (None, (8.63, 1.02), [(250, 7.96, 11.62), (655, 6.68, 10.37)]),
}


def ic_package(family, n):
    """λ0 = e^-a × Np^b for each stress (p. 125)."""
    rh, tc, rows = IC_PACKAGES[family]
    l_rh = math.exp(-rh[0]) * n ** rh[1] if rh else 0
    l_case = math.exp(-tc[0]) * n ** tc[1]
    for maxp, a_solder, a_mech in rows:
        if n <= maxp:
            return l_rh, l_case, math.exp(-a_solder) * n ** 0.92, math.exp(-a_mech) * n ** 0.92
    raise ValueError("%s: %d pins out of range" % (family, n))


# Integrated circuit chip λ0TH (p. 128)
IC_CHIPS = {
    "fpga": 0.076, "analog": 0.086, "mixed": 0.086, "microcontroller": 0.075,
    "flash": 0.060, "sram": 0.053, "dram": 0.047, "digital": 0.021,
}


def ic(tech, family, r, phases):
    """Integrated circuit (pp. 123-129)."""
    l_rh, l_case, l_solder, l_mech = ic_package(family, r["np"])
    l0th = IC_CHIPS[tech]

    def physical(ph):
        th = arrhenius(0.7, ph["tamb"] + r["t"]) if ph["on"] else 0
        rh = 0 if ph["on"] else l_rh * pi_rh(0.9, ph)
        return (l0th * th + l_case * pi_tcy_case(ph) + l_solder * pi_tcy_solder(ph)
                + rh + l_mech * pi_mech(ph))

    return fit(phases, physical, cs(10, 2, 1), r["placement"], PM_ACTIVE)


def opto(kind, family, r, phases):
    """Optocoupler (pp. 144-145), in an integrated circuit package."""
    l_rh, l_case, l_solder, l_mech = ic_package(family, r["np"])
    l0th, l_tcy_chip, l_mech_chip = (0.05, 0.01, 0.005) if kind == "photodiode" else (0.11, 0.021, 0.011)
    n = math.sqrt(max(r["n"], 1))
    l0th, l_tcy_chip, l_mech_chip = l0th * n, l_tcy_chip * n, l_mech_chip * n

    def physical(ph):
        th = arrhenius(0.4, ph["tamb"] + r["t"]) if ph["on"] else 0
        rh = 0 if ph["on"] else l_rh * pi_rh(0.9, ph)
        return (l0th * th + l_case * pi_tcy_case(ph) + (l_solder + l_tcy_chip) * pi_tcy_solder(ph)
                + rh + (l_mech + l_mech_chip) * pi_mech(ph))

    return fit(phases, physical, cs(7, 2, 2), r["placement"], PM_ACTIVE)


# Quartz crystals (pp. 163-164): λ0, γTH-EL, γTCy, γMech, γRH, (EOS, MOS, TOS), oscillator
PIEZO = {
    "res_tht": (0.82, 0.16, 0.46, 0.27, 0.11, (2, 10, 5), False),
    "res_smd": (0.79, 0.16, 0.59, 0.15, 0.10, (2, 10, 5), False),
    "osc_tht": (1.6, 0.32, 0.42, 0.14, 0.12, (7, 9, 3), True),
    "osc_smd": (1.63, 0.31, 0.53, 0.07, 0.09, (7, 9, 3), True),
}


def piezo(row, r, phases):
    l0, g_th, g_tcy, g_mech, g_rh, sens, osc = PIEZO[row]
    rating_el = 5 if osc and r["i"] >= 0.8 * r["imax"] else 1

    def physical(ph):
        th = 0
        if ph["on"]:
            rating_th = 5 if ph["tamb"] >= r["tmax"] - 40 else 1
            th = g_th * rating_th * rating_el
        rh = 0 if ph["on"] else g_rh * pi_rh(0.9, ph)
        return l0 * (th + g_tcy * pi_tcy_solder(ph) + g_mech * pi_mech(ph) + rh)

    return fit(phases, physical, cs(*sens), r["placement"], PM_PASSIVE)


def pcb(r, phases):
    """Printed circuit board (pp. 174-176); placement 1."""
    l0 = 5e-4 * r["layers"] ** 0.5 * r["mounts"] / 2 * r["piclass"] * r["pitechno"]

    def physical(ph):
        tv = 1 if ph["tamb"] < 110 else math.exp(0.2 * (ph["tamb"] - 110))
        return l0 * (0.6 * tv * pi_tcy_solder(ph) + 0.2 * tv * pi_mech(ph) + 0.18 * tv * pi_rh(0.8, ph)
                     + 0.02 * tv * ph["sal"] * ph["envir"] * ph["zone"] * ph["prot"])

    return fit(phases, physical, cs(4, 10, 8), 1.0, PM_PASSIVE)


# Connector mounting factors (p. 178)
MOUNT = {"pressfit": 1, "tht": 6, "smd": 10}


def connector(mount, r, phases):
    """Connector for printed circuits, fewer than one connection cycle a year
    (pp. 177-180); placement 1."""
    l0 = 0.1 * MOUNT[mount] * r["np"] ** 0.5 * 0.2

    def physical(ph):
        th = 0.58 * arrhenius(0.1, ph["tamb"] + r["t"]) if ph["on"] else 0
        chem = 0.20 * ph["sal"] * ph["envir"] * ph["zone"] * ph["prot"]
        return l0 * (th + 0.04 * pi_tcy_solder(ph) + 0.05 * pi_mech(ph) + 0.13 * pi_rh(0.8, ph) + chem)

    return fit(phases, physical, cs(1, 10, 3), 1.0, PM_PASSIVE)


def reference(r, phases):
    m = r["model"].split("/")
    if m[0] == "resistor":
        return resistor(m[1], r, phases)
    if m[0] in ("ceramic", "alu", "tant"):
        return capacitor(r["model"], r, phases)
    if m[0] == "inductor":
        return inductor(m[1], r, phases)
    if m[0] == "discrete":
        return discrete(m[1], m[2], r, phases)
    if m[0] == "power":
        return discrete("power_" + m[1], m[2], r, phases, power=True)
    if m[0] == "led":
        return led(m[1], m[2], r, phases)
    if m[0] == "ic":
        return ic(m[1], m[2], r, phases)
    if m[0] == "opto":
        return opto(m[1], m[2], r, phases)
    if m[0] == "piezo":
        return piezo(m[1], r, phases)
    if m[0] == "pcb":
        return pcb(r, phases)
    if m[0] == "connector":
        return connector(m[1], r, phases)
    raise ValueError("unknown model " + r["model"])


NUMBERS = ("placement", "value", "n", "np", "v", "vmax", "i", "imax", "p", "pmax", "t", "tmax",
           "layers", "mounts", "piclass", "pitechno")


def main():
    missions = {}
    w = csv.writer(sys.stdout)
    w.writerow(["name", "fit"])
    with open(os.path.join(HERE, "cases.csv")) as f:
        for row in csv.DictReader(f):
            r = dict(row)
            for k in NUMBERS:
                r[k] = float(row[k]) if row[k] else 0.0
            path = os.path.join(TESTDATA, row["mission"])
            if path not in missions:
                missions[path] = mission(path)
            w.writerow([row["name"], "%.12g" % reference(r, missions[path])])


if __name__ == "__main__":
    main()
