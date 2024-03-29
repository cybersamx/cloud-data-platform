# DBT

## Setup

### Dbt Installation

For the Mac, there are 2 ways to install dbt:

* Using pip - universal way to install and if you use multiple versions of Python (or has pyenv installed)
* Using homebrew - Mac-specific way to managing system-wide packages

Here are the instructions for installing dbt using pip since it's the most universal installation.

1. If you have `pyenv` installed, run `pyenv` to select the python version you want to use. For example:

   ```shell
   pyenv global 3.10.7  # Assume we want to use 3.10.7. 
   ```
   
1. Install `dbt-snowflake`. This will also install dbt, the plug-ins, and all the dependencies we need.

   ```shell
   pip install --user dbt-snowflake dbt-postgres
   ```
   
1. Verify that dbt has been successfully installed.

   ```shell
   $ dbt --version
   Core:
   - installed: 1.4.5
   - latest:    1.4.5 - Up to date!
   
   Plugins:
   - snowflake: 1.4.1 - Up to date!
   - postgres:  1.4.5 - Up to date!
   ```

> **Upgrade Notes**
> 
> To upgrade dbt run the following command: `pip install --upgrade dbt-core`

### Dbt Configuration

1. Create a profiles file in `~/.dbt`

   ```shell
   mkdir -p ~/.dbt
   cd ~/.dbt
   ```

1. Copy the `.env-example` to `.env`, which contains sensitive configuration values. Edit the configuration values in `.env`.

   ```shell
   cp .env-example .env
   vi .env
   ```

1. Source the .env to export the environment variables so that the sensitive configurations can be injected into the dbt profile.

   ```shell
   export $(cat .env | xargs)
   ```

1. We can use the `debug` command in dbt to check if our profiles settings are valid by actually connecting to the database.

   ```shell
   dbt debug
   ```

1. Run dbt to build the views/tables needed for data visualization.

   ```shell
   dbt build
   ```

## References

* [DBT Installation Overview](https://docs.getdbt.com/docs/get-started/installation)
* [DBT-Postgres Setup](https://docs.getdbt.com/reference/warehouse-setups/postgres-setup)
* [DBT-Snowflake Setup](https://docs.getdbt.com/reference/warehouse-setups/snowflake-setup)
