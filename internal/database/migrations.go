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
		`create table if not exists event_budget_items (
			id uuid primary key,
			event_id uuid not null references events(id) on delete cascade,
			description text not null,
			category text not null default '',
			amount double precision not null default 0,
			paid boolean not null default false,
			created_at timestamptz not null default now(),
			updated_at timestamptz not null default now()
		)`,
		`create table if not exists event_gifts (
			id uuid primary key,
			event_id uuid not null references events(id) on delete cascade,
			name text not null,
			description text not null default '',
			price double precision not null default 0,
			url text not null default '',
			reserved boolean not null default false,
			reserved_by text not null default '',
			created_at timestamptz not null default now(),
			updated_at timestamptz not null default now()
		)`,
		`alter table events add column if not exists tenant_id uuid`,
		`alter table events add column if not exists name text`,
		`alter table events add column if not exists title text`,
		`alter table events add column if not exists description text not null default ''`,
		`alter table events add column if not exists status text not null default 'draft'`,
		`alter table events add column if not exists timezone text not null default 'America/Sao_Paulo'`,
		`alter table events add column if not exists starts_at timestamptz`,
		`alter table events add column if not exists ends_at timestamptz`,
		`alter table events add column if not exists location text not null default ''`,
		`alter table events add column if not exists venue_name text`,
		`alter table events add column if not exists address text`,
		`alter table events add column if not exists spotify_playlist_url text`,
		`alter table events add column if not exists hero json not null default '{}'::json`,
		`alter table events add column if not exists content json not null default '{}'::json`,
		`alter table events add column if not exists theme json not null default '{}'::json`,
		`alter table events add column if not exists gallery json`,
		`alter table events add column if not exists seo json`,
		`alter table events add column if not exists capacity integer`,
		`alter table events add column if not exists slug text`,
		`alter table events add column if not exists created_at timestamptz not null default now()`,
		`alter table events add column if not exists updated_at timestamptz not null default now()`,
		`alter table guests add column if not exists event_id uuid`,
		`alter table guests add column if not exists name text`,
		`alter table guests add column if not exists email text`,
		`alter table guests add column if not exists phone text not null default ''`,
		`alter table guests add column if not exists status text not null default 'invited'`,
		`alter table guests add column if not exists created_at timestamptz not null default now()`,
		`alter table guests add column if not exists updated_at timestamptz not null default now()`,
		`alter table rsvps add column if not exists guest_id uuid`,
		`alter table rsvps add column if not exists event_id uuid`,
		`alter table rsvps add column if not exists status text`,
		`alter table rsvps add column if not exists created_at timestamptz not null default now()`,
		`alter table rsvps add column if not exists updated_at timestamptz not null default now()`,
		`alter table event_budget_items add column if not exists event_id uuid`,
		`alter table event_budget_items add column if not exists description text`,
		`alter table event_budget_items add column if not exists category text not null default ''`,
		`alter table event_budget_items add column if not exists amount double precision not null default 0`,
		`alter table event_budget_items add column if not exists paid boolean not null default false`,
		`alter table event_budget_items add column if not exists created_at timestamptz not null default now()`,
		`alter table event_budget_items add column if not exists updated_at timestamptz not null default now()`,
		`alter table event_gifts add column if not exists event_id uuid`,
		`alter table event_gifts add column if not exists name text`,
		`alter table event_gifts add column if not exists description text not null default ''`,
		`alter table event_gifts add column if not exists price double precision not null default 0`,
		`alter table event_gifts add column if not exists url text not null default ''`,
		`alter table event_gifts add column if not exists reserved boolean not null default false`,
		`alter table event_gifts add column if not exists reserved_by text not null default ''`,
		`alter table event_gifts add column if not exists created_at timestamptz not null default now()`,
		`alter table event_gifts add column if not exists updated_at timestamptz not null default now()`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}

	return nil
}
