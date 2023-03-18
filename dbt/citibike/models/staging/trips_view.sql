-- Build a view
select *
from {{ ref('trips') }}
