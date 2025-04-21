CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    emails TEXT[] NOT NULL DEFAULT '{}'
);

-- Table to store vocabulary words (Spanish and English)
CREATE TABLE IF NOT EXISTS words (
    id SERIAL PRIMARY KEY,
    spanish VARCHAR(255) NOT NULL,
    english VARCHAR(255) NOT NULL
);

-- Table to store lessons, each associated with a list of word IDs
CREATE TABLE IF NOT EXISTS lesson (
    id SERIAL PRIMARY KEY,
    word_list INTEGER[] NOT NULL
);

-- Table to store history of lesson attempts
-- CREATE TABLE IF NOT EXISTS history (
--     id SERIAL PRIMARY KEY,
--     lesson_id INTEGER NOT NULL,
--     user_id INTEGER NOT NULL,
--     results JSONB NOT NULL,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     CONSTRAINT fk_lesson FOREIGN KEY (lesson_id) REFERENCES lesson(id),
--     CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)
-- );