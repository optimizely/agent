### API acceptance tests for Optimizely Agent project

It is recommended to do it all in a virtual environment. 
  
1. Make sure you have latest Agent in the root directory. If not, then clone it:   
`git clone git@github.com:optimizely/agent.git`  
2. Install [Golang](https://golang.org/doc/install) (>= 1.13) if you don't have it already
3. `pip install -e .` 
4. `pytest -v agent_acceptance/test_acceptance/ --host http://localhost:8080`

To run an individual test 
TBD








