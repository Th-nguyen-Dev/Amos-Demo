-- Create update timestamp function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create Q&A knowledge base table
CREATE TABLE qa_pairs (
    id UUID PRIMARY KEY,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create conversations table
CREATE TABLE conversations (
    id UUID PRIMARY KEY,
    title TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create messages table (OpenAI format)
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    
    -- Extracted fields for querying (OpenAI standard)
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant', 'tool', 'system')),
    content TEXT,
    tool_call_id TEXT,
    
    -- Complete message in OpenAI format
    raw_message JSONB NOT NULL,
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for qa_pairs
CREATE INDEX idx_qa_id_desc ON qa_pairs(id DESC);
CREATE INDEX idx_qa_fts ON qa_pairs 
    USING gin(to_tsvector('english', question || ' ' || answer));

-- Create indexes for conversations
CREATE INDEX idx_conv_id_desc ON conversations(id DESC);

-- Create indexes for messages
CREATE INDEX idx_messages_conv ON messages(conversation_id, id ASC);
CREATE INDEX idx_messages_role ON messages(role);
CREATE INDEX idx_messages_content ON messages 
    USING gin(to_tsvector('english', content));
CREATE INDEX idx_messages_raw ON messages USING gin(raw_message);

-- Create triggers for updated_at
CREATE TRIGGER update_qa_pairs_updated_at BEFORE UPDATE ON qa_pairs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_conversations_updated_at BEFORE UPDATE ON conversations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

