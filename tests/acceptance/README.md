### API acceptance tests for Optimizely Agent

First, do everything from the agent's root directory. 

Python version 3.7 or greater is required.
It is recommended to set up a python virtual environment.     
Activate virtual environment.  
Install requirements `pip install -r tests/acceptance/requirements.txt` 


Run tests  
1. `pytest -v tests/acceptance/test_acceptance/ --host http://localhost:8080`

`--host` can point to any URL where agent service is located

For a nicer output you can add additional flags and run this  
`pytest -vv -rA --diff-type=split tests/acceptance/test_acceptance/ --host http://localhost:8080`

To run an individual test  
TBD










