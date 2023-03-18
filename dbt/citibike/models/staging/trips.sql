-- Override the config set in dbt_project.yml.
-- {{ config(materialized='table') }}

select * from trips

-- Uncomment the line below to remove records with null `id` values
-- where id is not null
