# vibestation

This is an experiment to see how effective an AI-assisted rewrite of an existing codebase can be.

The prompt for the LLM to kickoff the experiment is:

- We will rewrite the code for github.com/brexhq/substation.
- It should be written in the same language as the original code (Go).
- It should use a similar design as the original code, specifically:
  - Use of channels for inter-process communication.
  - Use of mutexes for synchronization.
  - Always return the results of data processing to the calling application.
  - Data processing functions (transforms) should use the same data structures (messages).
  - Data validation functions (conditions) should use the same data structures (messages).
  - transforms and conditions should be in the public API.

For the first iteration, we need the rewrite to deliver:

- A transform function to decompress gzipped data.
- A transform function to split string data into lines (each line is a new message).
- A transform function to print data to the console (stdout).
- A minimal application that reads a file from local disk and runs the transform functions.
