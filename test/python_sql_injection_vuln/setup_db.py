import sqlite3

def setup_database():
    conn = sqlite3.connect('test.db')
    cursor = conn.cursor()
    
    cursor.execute('DROP TABLE IF EXISTS users')
    
    cursor.execute('''
        CREATE TABLE users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT NOT NULL UNIQUE,
            password TEXT NOT NULL,
            email TEXT,
            is_admin INTEGER DEFAULT 0
        )
    ''')
    
    users = [
        ('admin', 'supersecret123', 'admin@example.com', 1),
        ('john', 'password123', 'john@example.com', 0),
        ('jane', 'qwerty', 'jane@example.com', 0),
        ('bob', 'letmein', 'bob@example.com', 0),
    ]
    
    cursor.executemany(
        'INSERT INTO users (username, password, email, is_admin) VALUES (?, ?, ?, ?)',
        users
    )
    
    conn.commit()
    conn.close()
    print("Database created with sample users!")

if __name__ == '__main__':
    setup_database()
