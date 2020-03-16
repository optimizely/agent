# Contributing to Optimizely Agent

We welcome contributions and feedback! All contributors must sign our [Contributor License Agreement (CLA)](https://docs.google.com/a/optimizely.com/forms/d/e/1FAIpQLSf9cbouWptIpMgukAKZZOIAhafvjFCV8hS00XJLWQnWDFtwtA/viewform) to be eligible to contribute. Please read the [README](README.md) to set up your development environment, then read the guidelines below for information on submitting your code.

## Development process

1. Fork the repository and create your branch from master.
2. Make sure to add tests!
3. Ensure lint check and tests are passing: `make lint`, `make test`
4. Push a branch with your changes to GitHub.
5. Open a PR from your fork into the master branch of the original repo
6. A repository maintainer will review your pull request and, if all goes well, squash and merge it!

## Pull request acceptance criteria

**All code must have test coverage.** Changes in functionality should have accompanying unit tests. Bug fixes should have accompanying regression tests.

## License

All contributions are under the CLA mentioned above.

For this project, Optimizely uses the Apache 2.0 license, and so asks that by contributing your code, you agree to license your contribution under the terms of the [Apache License v2.0](http://www.apache.org/licenses/LICENSE-2.0).

Your contributions should also include the following header:
```
/****************************************************************************
 * Copyright YEAR, Optimizely, Inc. and contributors                        *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    http://www.apache.org/licenses/LICENSE-2.0                            *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

```

The YEAR above should be the year of the contribution. If work on the file has been done over multiple years, list each year in the section above. Example: Optimizely writes the file and releases it in 2018. No changes are made in 2019. Change made in 2020. YEAR should be “2018, 2020”.

## Contact

If you have questions, please contact developers@optimizely.com.
