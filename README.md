# circlog
A cli tool for retreiving CircleCI logs.

In order to retreive the logs for a single CircleCI step, 5 API requests are necessary:
- Get pipelines for project
- Get workflows for pipeline
- Get jobs for workflow
- Get steps for job
- Get logs for step

Obviously this is rather cumbersome, especially when the final request uses information from multiple other responses and would look something like the following when using this cli tool.

`circlog logs <project-name> -j <job-number> -s <step-number> -i <step-index> -a <allocation-id>`

# circlog TUI
A simple TUI solves this. `circlog <project-name>` allows easy browsing to the required logs. Pressing the `Enter` key at this point will result in the `circlog` command needed to grab these logs to be printed to the terminal. This command can then be used to retreive the logs and directly dump them into the terminal.