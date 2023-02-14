-- name: GetTodos :many
select *
from todo
order by rank asc;

-- name: AddTodo :exec
insert into todo (id,task,created,updated,complete,rank) 
 values (?,?,?,?,?,?);

 -- name: RenameTodo :exec
 update todo 
 set task= ?, updated=?
 where id=?;

 -- name: GetTodo :one
 select *
 from todo
 where id=?;


-- name: Empty :exec
delete from todo;

-- name: DeleteTodo :exec
delete from todo where id=?;

-- name: ToggleTodoComplete :exec
update todo 
set complete=(case complete when 'Y' then 'N' else 'Y' end)
where id=?;

-- name: UpdateRank :exec
update todo
set rank=?
where id=?;

-- name: Search :many
select *
from todo
where task like ?
order by rank asc;