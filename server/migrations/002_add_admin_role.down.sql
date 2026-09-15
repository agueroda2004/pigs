DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM users
        WHERE role = 'Admin'::user_role
    ) THEN
        RAISE EXCEPTION 'Cannot remove Admin role while users have that role';
    END IF;
END
$$;

ALTER TABLE users ALTER COLUMN role DROP DEFAULT;
ALTER TABLE users ALTER COLUMN role TYPE TEXT USING role::TEXT;
DROP TYPE user_role;
CREATE TYPE user_role AS ENUM ('User');
ALTER TABLE users ALTER COLUMN role TYPE user_role USING role::user_role;
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'User';
