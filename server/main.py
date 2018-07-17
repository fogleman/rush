from flask import Flask, jsonify, g

import random
import sqlite3

# config

DB_PATH = 'rush.db'
DB_ATTR = '_db'

# app

app = Flask(__name__)

# hooks

def get_db():
    db = getattr(g, DB_ATTR, None)
    if db is None:
        db = sqlite3.connect(DB_PATH)
        db.row_factory = sqlite3.Row
        setattr(g, DB_ATTR, db)
    return db

def query_db(query, args=(), one=False):
    cursor = get_db().execute(query, args)
    result = cursor.fetchall()
    cursor.close()
    return (result[0] if result else None) if one else result

def row_dict(row):
    return dict((k, row[k]) for k in row.keys())

@app.teardown_appcontext
def close_connection(exception):
    db = getattr(g, DB_ATTR, None)
    if db is not None:
        db.close()

# views

@app.route('/random.json')
def random_json():
    COUNTS = [2577412,2577411,2577403,2577227,2575473,2563823,2518414,2412682,2236460,1990153,1696046,1396300,1128432,907256,727035,577576,452694,349953,267487,202610,152245,113712,84358,62386,46004,33870,25117,18538,13786,10224,7757,5919,4458,3398,2537,1883,1395,1022,776,567,425,326,234,171,113,85,63,47,33,23,15,13,11,8,4,2,2,2,1,1]
    i = random.randint(15, 40) - 1
    rowid = random.randint(1, COUNTS[i])

    db = get_db()
    q = 'select * from rush where rowid = ?;'
    row = query_db(q, (rowid,), one=True)
    resp = jsonify(row_dict(row))
    resp.headers['Access-Control-Allow-Origin'] = '*'
    return resp

# main

if __name__ == '__main__':
    app.run(debug=True)
