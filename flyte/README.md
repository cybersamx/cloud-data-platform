# Flyte

Flyte is an open-source workflow orchestration platform that runs data pipelines and ML workloads. The execution of dbt models that we built will be orchestrated by Flyte.

## Setup

1. Install flyte cli.

   ```shell
   $ brew install flyteorg/homebrew-tap/flytectl
   ```

1. Create a virtual environment.

   ```shell
   $ python -m venv .venv
   $ source .venv/bin/activate
   (.venv) $ # Should see the virtual env name as a prefix
   ```

1. Install packages

   ```shell
   $ pip install -r requirements.txt
   ```

1. Launch Rancher or Docker Desktop to start a container engine. Run the following command to create a local k8s cluster. Port 6443 needs to be available for this to work, so disable k8s on Rancher or Docker Desktop and let flytectl enable it. 

   ```shell
   $ flytectl demo start
   ```

1. Run the workflow.

   ```shell
   $ pyflyte run --remote workflow.py training_workflow --hyperparameters '{"C": 0.1}'
   ```
   
1. Go to <http://localhost:30080> to check the result.

# Notes

* Flyte tasks are core building blocks of Flyte workflows and is defined as a Python function annotated with the `@task` decorator.
* Tasks are strongly typed. Typechecking allows some errors to be caught during compile-time.

## References

* [Flyte Home Page](https://flyte.org/)
* [Flyte Docs: Getting Started](https://docs.flyte.org/projects/cookbook/en/latest/index.html)
