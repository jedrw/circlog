# circlog
A cli tool for retreiving CircleCI logs.

In order to retreive the logs for a single CircleCI step, 5 API requests are necessary:
- Get pipelines for project - `circlog pipelines <project>`
- Get workflows for pipeline - `circlog workflows <project> -l <pipeline-id>`
- Get jobs for workflow - `circlog jobs <project> -w <workflow-id>`
- Get steps for job - `circlog steps <project> -j <job-number>`
- Get logs for step - `circlog logs <project-name> -j <job-number> -s <step-number> -i <step-index> -a <allocation-id>`

Obviously this is rather cumbersome, especially when the final request uses information gathered from multiple other responses.

# circlog TUI
A simple TUI solves this. `circlog <project-name>` allows easy browsing to the required logs. Pressing the `D` key at this point will result in the `circlog` command needed to grab these logs being printed to the terminal. This command can then be used to retreive the logs and directly dump them into the terminal.

## Configuration
If you have the CircleCi CLI tool installed and configured already circlog will work 'out of the box' by using the token set in the CircleCi CLI config file.

You may also add a token to the CIRCLECI_TOKEN env var which will be used instead.
