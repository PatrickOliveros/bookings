INSERT INTO public.rooms (room_name,created_at,updated_at) VALUES
	 ('Generals Quarters',NOW(), NOW()),
	 ('Majors Suite', NOW(),NOW())

	 ON CONFLICT ON ID DO NOTHING