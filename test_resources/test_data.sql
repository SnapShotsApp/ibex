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

create table if not exists watermarks(
       id integer primary key,
       photographer_info_id integer,
       disabled boolean,
       "default" boolean,
       logo varchar(255),
       alpha integer,
       "scale" integer,
       "offset" integer,
       "position" varchar(255)
);

insert into pictures values(1, 1, 1, 'test_pic.jpg');
insert into pictures values(2, 2, 1, 'guest_test_pic.jpg');

insert into events values(1, 1);
insert into events values(2, 3);

insert into photographer_infos values(1, 1, 'test_watermark.jpg');
insert into photographer_infos values(2, 3, 'extra_test_watermark.jpg');
insert into photographer_infos values(3, 4, null);
insert into photographer_infos values(4, 5, null);

insert into watermarks values(1, 1, FALSE, TRUE, 'test_watermark.jpg', 70, 40, 3, E'---\n- bottom\n- left\n');
insert into watermarks values(2, 1, FALSE, FALSE, 'test_watermark_error.jpg', 100, 100, 100, E'---\n- bottom\n');
insert into watermarks values(3, 2, FALSE, TRUE, 'test_watermark2.jpg', 100, 100, 1, E'---\n- top\n- left\n');
insert into watermarks values(4, 3, FALSE, TRUE, 'test_watermark3.jpg', 20, 75, 0, E'---\n- top\n- right\n');
insert into watermarks values(5, 4, TRUE, TRUE, null, null, null, null, null);
