-- name: picture_by_id
select user_id, event_id, attachment from pictures where id = $1

-- name: event_by_id
select owner_id from events where id = $1

-- name: photographer_info_by_user_id
select id, picture from photographer_infos where user_id = $1

-- name: watermark_by_photographer_info_id
select watermarks.id, watermarks.logo, watermarks.disabled, watermarks.alpha,
watermarks.scale, watermarks.offset, watermarks.position
from watermarks
where photographer_info_id = $1 and watermarks.default = 't'
limit 1

