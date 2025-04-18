INSERT INTO users (id, user_name, created_at, updated_at)
VALUES
  ('11111111-1111-1111-1111-111111111111'::UUID, 'alice', now(), now()),
  ('22222222-2222-2222-2222-222222222222'::UUID, 'bob', now(), now()),
  ('33333333-3333-3333-3333-333333333333'::UUID, 'carol', now(), now()),
  ('44444444-4444-4444-4444-444444444444'::UUID, 'dave', now(), now());

INSERT INTO follows (follower_id, followee_id)
VALUES
  ('11111111-1111-1111-1111-111111111111'::UUID, '22222222-2222-2222-2222-222222222222'::UUID), -- alice follows bob
  ('11111111-1111-1111-1111-111111111111'::UUID, '33333333-3333-3333-3333-333333333333'::UUID), -- alice follows carol
  ('22222222-2222-2222-2222-222222222222'::UUID, '33333333-3333-3333-3333-333333333333'::UUID), -- bob follows carol
  ('33333333-3333-3333-3333-333333333333'::UUID, '44444444-4444-4444-4444-444444444444'::UUID); -- carol follows dave

INSERT INTO posts (id, user_id, content, created_at, updated_at)
VALUES
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::UUID, '22222222-2222-2222-2222-222222222222'::UUID, 'Hello from Bob!', now() - INTERVAL '5 days', now() - INTERVAL '5 days'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::UUID, '33333333-3333-3333-3333-333333333333'::UUID, 'Carol here, nice to meet you!', now() - INTERVAL '4 days', now() - INTERVAL '4 days'),
  ('cccccccc-cccc-cccc-cccc-cccccccccccc'::UUID, '44444444-4444-4444-4444-444444444444'::UUID, 'Dave just joined!', now() - INTERVAL '3 days', now() - INTERVAL '3 days'),
  ('dddddddd-dddd-dddd-dddd-dddddddddddd'::UUID, '33333333-3333-3333-3333-333333333333'::UUID, 'Another post from Carol', now() - INTERVAL '2 days', now() - INTERVAL '2 days'),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee'::UUID, '22222222-2222-2222-2222-222222222222'::UUID, 'Bob again!', now() - INTERVAL '1 day', now() - INTERVAL '1 day');
