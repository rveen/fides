# FIDES reliability library for Go

This library is based on FIDES edition 2022. It is not a comprehensive implementation (see the notes below).

Although FIDES and all reliability methods derived from or similar to MIL-STD-217 are currently not accurate
according to reports from different sources, NASA included (see the references), some kind of failure estimation
for electronic devices is necessary. FIDES is considered here as a starting point, and as models evolve,
eventually also this library will do.

## Install

    # git clone https://github.com/rveen/fides
    # cd fides/cmd/fides
    # go build
    # ./fides bom.csv db.csv mission.csv

This will do a FIT calculation on the sample BOM provided and print it on screen.

## Notes on this implementation

- Missing components:
  - COTS
- Process factors are mostly ignored (they are considered subjective):
  - PiRuggedized is set to the default value 1.7
  - PiLF is ignored (=1)
  - PiPM

## references

- http://fides-reliability.org/
- https://www.sciencedirect.com/science/article/pii/S2772671123002486
- https://nepp.nasa.gov/docs/etw/2019/0619WED/1630%20-%20Bourbouse%20-%20NEPP_2019_FIDES.pdf
- https://calce.umd.edu/event/18917/elseviers-e-prime-presentation-by-dr-diganta-das-assessment-of-the-fides-guide-2022
- https://ieeexplore.ieee.org/document/9319134


