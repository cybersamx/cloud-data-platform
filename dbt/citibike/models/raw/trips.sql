-- Override the config set in dbt_project.yml.
-- {{ config(materialized='table') }}

-- Make a view of the raw data with data types.

with trips as (
  select * from {{ ref('raw_trips') }}
)

select * from trips

-- Uncomment the line below to remove records with null `id` values
-- where id is not null
