create table if not exists pictures(
  id integer primary key,
  user_id integer,
  event_id integer,
  attachment varchar(255)
);

create table if not exists events(
       id integer primary key,
       owner_id integer
);

create table if not exists photographer_infos(
       id integer primary key,
       user_id integer,
       picture varchar(255)
);

insert into pictures values(1, 1, 1, 'test_pic.jpg');
insert into pictures values(2, 2, 1, 'guest_test_pic.jpg');

insert into events values(1, 1);
insert into events values(2, 3);

insert into photographer_infos values(1, 1, 'test_watermark.jpg');
insert into photographer_infos values(2, 3, 'extra_test_watermark.jpg');
insert into photographer_infos values(3, 4, null);

