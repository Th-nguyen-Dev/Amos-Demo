-- Seed data for demo and testing purposes

-- Insert demo QA pairs
INSERT INTO qa_pairs (id, question, answer) VALUES
    (gen_random_uuid(), 'What is your refund policy?', 'We offer a 30-day money-back guarantee on all purchases. If you are not satisfied with your purchase, you can request a full refund within 30 days of the purchase date.'),
    (gen_random_uuid(), 'What are your shipping options?', 'We offer standard shipping (5-7 business days) for free on orders over $50. Express shipping (2-3 business days) is available for $9.99, and overnight shipping is available for $19.99.'),
    (gen_random_uuid(), 'How can I track my order?', 'Once your order ships, you will receive a tracking number via email. You can use this tracking number on our website or the carrier''s website to track your package in real-time.'),
    (gen_random_uuid(), 'Do you ship internationally?', 'Yes, we ship to over 50 countries worldwide. International shipping rates and delivery times vary by destination. Please check our shipping page for more details.'),
    (gen_random_uuid(), 'What payment methods do you accept?', 'We accept all major credit cards (Visa, MasterCard, American Express, Discover), PayPal, Apple Pay, and Google Pay. All transactions are secure and encrypted.'),
    (gen_random_uuid(), 'How do I reset my password?', 'Click on the "Forgot Password" link on the login page. Enter your email address, and we will send you a link to reset your password. The link is valid for 24 hours.'),
    (gen_random_uuid(), 'Can I cancel or modify my order?', 'You can cancel or modify your order within 1 hour of placing it by contacting our customer service team. After that, the order may have already been processed for shipping.'),
    (gen_random_uuid(), 'What is your customer support availability?', 'Our customer support team is available Monday through Friday, 9 AM to 6 PM EST. You can reach us via email at support@example.com or through our live chat on the website.'),
    (gen_random_uuid(), 'Do you offer gift wrapping?', 'Yes, we offer complimentary gift wrapping for all orders. Simply select the gift wrapping option at checkout, and you can include a personalized message.'),
    (gen_random_uuid(), 'What is your exchange policy?', 'We accept exchanges within 30 days of purchase for items in new, unused condition with original tags attached. Please contact customer service to initiate an exchange.');

-- Insert demo conversations
INSERT INTO conversations (id, title) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Customer Inquiry: Shipping Question'),
    ('00000000-0000-0000-0000-000000000002', 'Support Request: Order Tracking'),
    ('00000000-0000-0000-0000-000000000003', 'General Question: Payment Methods');

-- Insert demo messages for conversation 1 (Shipping Question)
INSERT INTO messages (id, conversation_id, role, content, tool_call_id, raw_message) VALUES
    (
        gen_random_uuid(),
        '00000000-0000-0000-0000-000000000001',
        'user',
        'Hi, I was wondering about your international shipping options?',
        NULL,
        '{"role": "user", "content": "Hi, I was wondering about your international shipping options?"}'::jsonb
    ),
    (
        gen_random_uuid(),
        '00000000-0000-0000-0000-000000000001',
        'assistant',
        'Hello! I''d be happy to help you with information about our international shipping. We ship to over 50 countries worldwide. Would you like to know about specific countries or general rates?',
        NULL,
        '{"role": "assistant", "content": "Hello! I''d be happy to help you with information about our international shipping. We ship to over 50 countries worldwide. Would you like to know about specific countries or general rates?"}'::jsonb
    ),
    (
        gen_random_uuid(),
        '00000000-0000-0000-0000-000000000001',
        'user',
        'What about shipping to Canada?',
        NULL,
        '{"role": "user", "content": "What about shipping to Canada?"}'::jsonb
    ),
    (
        gen_random_uuid(),
        '00000000-0000-0000-0000-000000000001',
        'assistant',
        'Shipping to Canada typically takes 7-10 business days. The cost varies by weight and destination, but generally ranges from $12-$25 USD. Free shipping is available on Canadian orders over $100.',
        NULL,
        '{"role": "assistant", "content": "Shipping to Canada typically takes 7-10 business days. The cost varies by weight and destination, but generally ranges from $12-$25 USD. Free shipping is available on Canadian orders over $100."}'::jsonb
    );

-- Insert demo messages for conversation 2 (Order Tracking)
INSERT INTO messages (id, conversation_id, role, content, tool_call_id, raw_message) VALUES
    (
        gen_random_uuid(),
        '00000000-0000-0000-0000-000000000002',
        'user',
        'I placed an order 3 days ago but haven''t received a tracking number yet.',
        NULL,
        '{"role": "user", "content": "I placed an order 3 days ago but haven''t received a tracking number yet."}'::jsonb
    ),
    (
        gen_random_uuid(),
        '00000000-0000-0000-0000-000000000002',
        'assistant',
        'I apologize for the delay. Let me look into this for you. Could you please provide your order number?',
        NULL,
        '{"role": "assistant", "content": "I apologize for the delay. Let me look into this for you. Could you please provide your order number?"}'::jsonb
    ),
    (
        gen_random_uuid(),
        '00000000-0000-0000-0000-000000000002',
        'user',
        'Sure, it''s #ORD-12345',
        NULL,
        '{"role": "user", "content": "Sure, it''s #ORD-12345"}'::jsonb
    ),
    (
        gen_random_uuid(),
        '00000000-0000-0000-0000-000000000002',
        'assistant',
        'Thank you! I''ve checked your order #ORD-12345. Your package was shipped this morning, and you should receive a tracking email within the next hour. Your tracking number is 1Z999AA10123456784.',
        NULL,
        '{"role": "assistant", "content": "Thank you! I''ve checked your order #ORD-12345. Your package was shipped this morning, and you should receive a tracking email within the next hour. Your tracking number is 1Z999AA10123456784.", "tool_calls": [{"type": "function", "function": {"name": "lookup_order", "arguments": "{\"order_id\": \"ORD-12345\"}"}}]}'::jsonb
    );

-- Insert demo messages for conversation 3 (Payment Methods)
INSERT INTO messages (id, conversation_id, role, content, tool_call_id, raw_message) VALUES
    (
        gen_random_uuid(),
        '00000000-0000-0000-0000-000000000003',
        'user',
        'Do you accept cryptocurrency payments?',
        NULL,
        '{"role": "user", "content": "Do you accept cryptocurrency payments?"}'::jsonb
    ),
    (
        gen_random_uuid(),
        '00000000-0000-0000-0000-000000000003',
        'assistant',
        'Currently, we accept all major credit cards, PayPal, Apple Pay, and Google Pay. We don''t accept cryptocurrency at this time, but we''re always evaluating new payment options based on customer demand.',
        NULL,
        '{"role": "assistant", "content": "Currently, we accept all major credit cards, PayPal, Apple Pay, and Google Pay. We don''t accept cryptocurrency at this time, but we''re always evaluating new payment options based on customer demand."}'::jsonb
    );

