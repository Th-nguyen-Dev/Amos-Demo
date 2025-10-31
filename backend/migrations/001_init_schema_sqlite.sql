-- Create Q&A knowledge base table
CREATE TABLE qa_pairs (
    id TEXT PRIMARY KEY,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create conversations table
CREATE TABLE conversations (
    id TEXT PRIMARY KEY,
    title TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create messages table (OpenAI format)
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL,
    
    -- Extracted fields for querying (OpenAI standard)
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant', 'tool', 'system')),
    content TEXT,
    tool_call_id TEXT,
    
    -- Complete message in OpenAI format (stored as JSON text)
    raw_message TEXT NOT NULL,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);

-- Create indexes for qa_pairs
CREATE INDEX idx_qa_created_at ON qa_pairs(created_at DESC);
CREATE INDEX idx_qa_question ON qa_pairs(question);

-- Create indexes for conversations
CREATE INDEX idx_conv_created_at ON conversations(created_at DESC);

-- Create indexes for messages
CREATE INDEX idx_messages_conv ON messages(conversation_id, created_at ASC);
CREATE INDEX idx_messages_role ON messages(role);

-- Create triggers for updated_at
CREATE TRIGGER update_qa_pairs_updated_at 
AFTER UPDATE ON qa_pairs
BEGIN
    UPDATE qa_pairs SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER update_conversations_updated_at 
AFTER UPDATE ON conversations
BEGIN
    UPDATE conversations SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

