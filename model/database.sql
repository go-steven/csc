drop table if exists `csc_users`;

CREATE TABLE `csc_users` (
  `uid` int(10) unsigned NOT NULL,
  `name` varchar(64) DEFAULT NULL,
  `nick` varchar(64) DEFAULT NULL,
  `affiliate` varchar(10) DEFAULT NULL,
  `role` int(1) DEFAULT 0,
  `manager` int(1) DEFAULT 0,
  `signin_at` datetime DEFAULT NULL,  
  `signup_at` datetime DEFAULT NULL,
  `status` int(1) DEFAULT 0,
  `session` varchar(255) default null,
  `kefu_id` int(10) unsigned NOT NULL default 0,
  PRIMARY KEY (`uid`, `role`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT;

drop table if exists `csc_msgs`;
CREATE TABLE `csc_msgs` (
  `id` varchar(255) NOT NULL,
  `visitor_id` int(10) unsigned NOT NULL,
  `kefu_id` int(10) unsigned NOT NULL,
  `initiator` int(1) unsigned NOT NULL,
  `status` int(1) unsigned NOT NULL,
  `create_time` datetime NOT NULL,
  `send_time` datetime DEFAULT NULL,
  `content` text,
  `msg_type` int(3) unsigned DEFAULT '1',
  `session_id` varchar(255) DEFAULT NULL,
  `response_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `visitor_id` (`visitor_id`),
  KEY `kefu_id` (`kefu_id`),
  KEY `create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT;

drop table if exists `csc_kefu_stats`;
CREATE TABLE `csc_kefu_stats` (
  `kefu_id` int(10) unsigned NOT NULL,
  `stat_date` varchar(255) NOT NULL,
  `total_visitors` int(10) unsigned DEFAULT NULL,
  `dup_visitors` int(10) unsigned DEFAULT NULL,
  `no_response_visitors` int(10) unsigned DEFAULT NULL,
  `visitor_end` int(10) unsigned DEFAULT NULL,
  `avg_duration` int(10) unsigned DEFAULT NULL,
  `avg_response_duration` int(10) unsigned DEFAULT NULL,
  `avg_visitor_content` int(10) unsigned DEFAULT NULL,
  `avg_kefu_content` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`kefu_id`,`stat_date`),
  KEY `stat_date` (`stat_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT;
