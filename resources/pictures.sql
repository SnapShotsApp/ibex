-- name: picture_by_id
select user_id, event_id, attachment from pictures where id = $1

-- name: event_by_id
select owner_id from events where id = $1

-- name: photographer_info_by_user_id
select id, picture from photographer_infos where user_id = $1
