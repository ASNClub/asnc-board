-- rollback: drop all tables in reverse order

DROP TABLE IF EXISTS activity_events;
DROP TABLE IF EXISTS notification_settings;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS payment_events;
DROP TABLE IF EXISTS sponsorships;
DROP TABLE IF EXISTS comment_votes;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS post_votes;
DROP TABLE IF EXISTS post_media;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS community_bans;
DROP TABLE IF EXISTS community_stars;
DROP TABLE IF EXISTS community_follows;
DROP TABLE IF EXISTS community_tags;
DROP TABLE IF EXISTS communities;
DROP TABLE IF EXISTS pinned_repos;
DROP TABLE IF EXISTS user_git_accounts;
DROP TABLE IF EXISTS friendships;
DROP TABLE IF EXISTS user_follows;
DROP TABLE IF EXISTS user_platforms;
DROP TABLE IF EXISTS user_tags;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "uuid-ossp";
