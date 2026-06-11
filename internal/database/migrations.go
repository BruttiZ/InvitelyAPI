package database

import "database/sql"

func RunMigrations(db *sql.DB) error {
	if err := db.Ping(); err != nil {
		return err
	}

	statements := []string{
		`create table if not exists tenants (
			id uuid primary key,
			name text not null,
			slug text not null unique,
			created_at timestamptz not null default now(),
			updated_at timestamptz not null default now()
		)`,
		`create table if not exists users (
			id uuid primary key,
			tenant_id uuid null references tenants(id) on delete set null,
			supabase_user_id uuid null unique,
			email text not null unique,
			name text not null,
			role text not null default 'owner',
			created_at timestamptz not null default now(),
			updated_at timestamptz not null default now()
		)`,
		`create table if not exists events (
			id uuid primary key,
			tenant_id uuid not null references tenants(id) on delete cascade,
			title text not null,
			description text not null default '',
			starts_at timestamptz not null,
			ends_at timestamptz null,
			location text not null default '',
			slug text not null,
			created_at timestamptz not null default now(),
			updated_at timestamptz not null default now(),
			unique (tenant_id, slug)
		)`,
		`create table if not exists guests (
			id uuid primary key,
			event_id uuid not null references events(id) on delete cascade,
			name text not null,
			email text not null,
			phone text not null default '',
			status text not null default 'invited',
			created_at timestamptz not null default now(),
			updated_at timestamptz not null default now(),
			unique (event_id, email)
		)`,
		`create table if not exists rsvps (
			id uuid primary key,
			guest_id uuid not null references guests(id) on delete cascade,
			event_id uuid not null references events(id) on delete cascade,
			status text not null,
			created_at timestamptz not null default now(),
			updated_at timestamptz not null default now(),
			unique (guest_id)
		)`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}

	return nil
}
