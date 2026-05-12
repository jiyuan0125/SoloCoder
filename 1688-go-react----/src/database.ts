import * as sqlite3 from 'sqlite3';
import * as path from 'path';
import { Scale, Scale as ScaleType } from './types';
import { v4 as uuidv4 } from 'uuid';

const DB_PATH = path.join(__dirname, '..', 'assessment.db');

export const db = new sqlite3.Database(DB_PATH);

export const initDatabase = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    db.serialize(() => {
      db.run(`
        CREATE TABLE IF NOT EXISTS assessments (
          id TEXT PRIMARY KEY,
          user_id TEXT NOT NULL,
          scale_id TEXT NOT NULL,
          start_time INTEGER NOT NULL,
          status TEXT NOT NULL DEFAULT 'in_progress',
          answers TEXT NOT NULL DEFAULT '[]',
          last_active_time INTEGER NOT NULL,
          FOREIGN KEY (scale_id) REFERENCES scales(id)
        )
      `);

      db.run(`
        CREATE TABLE IF NOT EXISTS assessment_results (
          id TEXT PRIMARY KEY,
          assessment_id TEXT NOT NULL,
          user_id TEXT NOT NULL,
          scale_id TEXT NOT NULL,
          raw_score INTEGER NOT NULL,
          standard_score INTEGER NOT NULL,
          level TEXT NOT NULL,
          dimension_scores TEXT,
          completed_at INTEGER NOT NULL,
          report TEXT NOT NULL,
          FOREIGN KEY (assessment_id) REFERENCES assessments(id)
        )
      `);

      db.run(`
        CREATE TABLE IF NOT EXISTS alerts (
          id TEXT PRIMARY KEY,
          assessment_result_id TEXT NOT NULL,
          user_id TEXT NOT NULL,
          scale_id TEXT NOT NULL,
          level TEXT NOT NULL,
          priority TEXT NOT NULL DEFAULT 'normal',
          status TEXT NOT NULL DEFAULT 'unprocessed',
          notified_users TEXT NOT NULL DEFAULT '[]',
          created_at INTEGER NOT NULL,
          processing_record TEXT,
          resolved_at INTEGER,
          processed_by TEXT,
          FOREIGN KEY (assessment_result_id) REFERENCES assessment_results(id)
        )
      `);

      db.run(`
        CREATE TABLE IF NOT EXISTS notifications (
          id TEXT PRIMARY KEY,
          alert_id TEXT NOT NULL,
          user_id TEXT NOT NULL,
          recipient_id TEXT NOT NULL,
          content TEXT NOT NULL,
          sent_at INTEGER NOT NULL,
          read_at INTEGER,
          FOREIGN KEY (alert_id) REFERENCES alerts(id)
        )
      `);

      db.run(`
        CREATE TABLE IF NOT EXISTS user_relations (
          user_id TEXT PRIMARY KEY,
          supervisor_id TEXT,
          class_teacher_id TEXT
        )
      `);

      seedScales((err) => {
        if (err) reject(err);
        else resolve();
      });
    });
  });
};

const builtInScales: Scale[] = [
  {
    id: 'scl-90',
    name: 'SCL-90 症状自评量表',
    targetGroup: '16岁以上人群',
    scoringType: 'sum',
    questions: [
      {
        id: 'q1',
        text: '头痛',
        isReverseScored: false,
        dimension: '躯体化',
        options: [
          { id: 'o1', text: '没有', score: 0 },
          { id: 'o2', text: '很轻', score: 1 },
          { id: 'o3', text: '中度', score: 2 },
          { id: 'o4', text: '偏重', score: 3 },
          { id: 'o5', text: '严重', score: 4 }
        ]
      },
      {
        id: 'q2',
        text: '神经过敏，心中不踏实',
        isReverseScored: false,
        dimension: '焦虑',
        options: [
          { id: 'o1', text: '没有', score: 0 },
          { id: 'o2', text: '很轻', score: 1 },
          { id: 'o3', text: '中度', score: 2 },
          { id: 'o4', text: '偏重', score: 3 },
          { id: 'o5', text: '严重', score: 4 }
        ]
      },
      {
        id: 'q3',
        text: '头脑中有不必要的想法或字句盘旋',
        isReverseScored: false,
        dimension: '强迫',
        options: [
          { id: 'o1', text: '没有', score: 0 },
          { id: 'o2', text: '很轻', score: 1 },
          { id: 'o3', text: '中度', score: 2 },
          { id: 'o4', text: '偏重', score: 3 },
          { id: 'o5', text: '严重', score: 4 }
        ]
      },
      {
        id: 'q4',
        text: '对事物不感兴趣',
        isReverseScored: false,
        dimension: '抑郁',
        options: [
          { id: 'o1', text: '没有', score: 0 },
          { id: 'o2', text: '很轻', score: 1 },
          { id: 'o3', text: '中度', score: 2 },
          { id: 'o4', text: '偏重', score: 3 },
          { id: 'o5', text: '严重', score: 4 }
        ]
      },
      {
        id: 'q5',
        text: '感到前途没有希望',
        isReverseScored: false,
        dimension: '抑郁',
        options: [
          { id: 'o1', text: '没有', score: 0 },
          { id: 'o2', text: '很轻', score: 1 },
          { id: 'o3', text: '中度', score: 2 },
          { id: 'o4', text: '偏重', score: 3 },
          { id: 'o5', text: '严重', score: 4 }
        ]
      }
    ],
    normTable: [
      { rawScore: 0, standardScore: 40, level: 'normal' },
      { rawScore: 5, standardScore: 50, level: 'normal' },
      { rawScore: 10, standardScore: 60, level: 'mild' },
      { rawScore: 15, standardScore: 70, level: 'moderate' },
      { rawScore: 20, standardScore: 80, level: 'severe' }
    ]
  },
  {
    id: 'phq-9',
    name: 'PHQ-9 抑郁自评量表',
    targetGroup: '12岁以上人群',
    scoringType: 'sum',
    questions: [
      {
        id: 'q1',
        text: '做事提不起劲或没有兴趣',
        isReverseScored: false,
        options: [
          { id: 'o1', text: '完全不会', score: 0 },
          { id: 'o2', text: '几天', score: 1 },
          { id: 'o3', text: '一半以上的天数', score: 2 },
          { id: 'o4', text: '几乎每天', score: 3 }
        ]
      },
      {
        id: 'q2',
        text: '感到心情低落、沮丧或绝望',
        isReverseScored: false,
        options: [
          { id: 'o1', text: '完全不会', score: 0 },
          { id: 'o2', text: '几天', score: 1 },
          { id: 'o3', text: '一半以上的天数', score: 2 },
          { id: 'o4', text: '几乎每天', score: 3 }
        ]
      },
      {
        id: 'q3',
        text: '入睡困难、睡眠不安或睡眠过多',
        isReverseScored: false,
        options: [
          { id: 'o1', text: '完全不会', score: 0 },
          { id: 'o2', text: '几天', score: 1 },
          { id: 'o3', text: '一半以上的天数', score: 2 },
          { id: 'o4', text: '几乎每天', score: 3 }
        ]
      },
      {
        id: 'q4',
        text: '感到疲倦或没有活力',
        isReverseScored: false,
        options: [
          { id: 'o1', text: '完全不会', score: 0 },
          { id: 'o2', text: '几天', score: 1 },
          { id: 'o3', text: '一半以上的天数', score: 2 },
          { id: 'o4', text: '几乎每天', score: 3 }
        ]
      },
      {
        id: 'q5',
        text: '食欲不振或吃太多',
        isReverseScored: false,
        options: [
          { id: 'o1', text: '完全不会', score: 0 },
          { id: 'o2', text: '几天', score: 1 },
          { id: 'o3', text: '一半以上的天数', score: 2 },
          { id: 'o4', text: '几乎每天', score: 3 }
        ]
      },
      {
        id: 'q6',
        text: '觉得自己很糟，或觉得自己是个失败者',
        isReverseScored: false,
        options: [
          { id: 'o1', text: '完全不会', score: 0 },
          { id: 'o2', text: '几天', score: 1 },
          { id: 'o3', text: '一半以上的天数', score: 2 },
          { id: 'o4', text: '几乎每天', score: 3 }
        ]
      },
      {
        id: 'q7',
        text: '对事物专注有困难',
        isReverseScored: false,
        options: [
          { id: 'o1', text: '完全不会', score: 0 },
          { id: 'o2', text: '几天', score: 1 },
          { id: 'o3', text: '一半以上的天数', score: 2 },
          { id: 'o4', text: '几乎每天', score: 3 }
        ]
      },
      {
        id: 'q8',
        text: '动作或说话速度缓慢到别人可以察觉',
        isReverseScored: false,
        options: [
          { id: 'o1', text: '完全不会', score: 0 },
          { id: 'o2', text: '几天', score: 1 },
          { id: 'o3', text: '一半以上的天数', score: 2 },
          { id: 'o4', text: '几乎每天', score: 3 }
        ]
      },
      {
        id: 'q9',
        text: '有不如死掉或用某种方式伤害自己的念头',
        isReverseScored: false,
        options: [
          { id: 'o1', text: '完全不会', score: 0 },
          { id: 'o2', text: '几天', score: 1 },
          { id: 'o3', text: '一半以上的天数', score: 2 },
          { id: 'o4', text: '几乎每天', score: 3 }
        ]
      }
    ],
    normTable: [
      { rawScore: 0, standardScore: 40, level: 'normal' },
      { rawScore: 5, standardScore: 55, level: 'mild' },
      { rawScore: 10, standardScore: 70, level: 'moderate' },
      { rawScore: 20, standardScore: 90, level: 'severe' }
    ]
  }
];

const seedScales = (callback: (err: Error | null) => void) => {
  db.all("SELECT name FROM sqlite_master WHERE type='table' AND name='scales'", (err, rows) => {
    if (err) {
      callback(err);
      return;
    }

    if (rows.length === 0) {
      db.run(`
        CREATE TABLE scales (
          id TEXT PRIMARY KEY,
          name TEXT NOT NULL,
          target_group TEXT NOT NULL,
          data TEXT NOT NULL
        )
      `, (createErr) => {
        if (createErr) {
          callback(createErr);
          return;
        }

        let remaining = builtInScales.length;
        if (remaining === 0) {
          callback(null);
          return;
        }

        for (const scale of builtInScales) {
          db.run(
            'INSERT INTO scales (id, name, target_group, data) VALUES (?, ?, ?, ?)',
            [scale.id, scale.name, scale.targetGroup, JSON.stringify(scale)],
            (insertErr) => {
              remaining--;
              if (remaining === 0) {
                callback(insertErr || null);
              }
            }
          );
        }
      });
    } else {
      callback(null);
    }
  });
};

export const getScaleById = (id: string): Promise<ScaleType | null> => {
  return new Promise((resolve, reject) => {
    db.get('SELECT data FROM scales WHERE id = ?', [id], (err, row: any) => {
      if (err) reject(err);
      else resolve(row ? JSON.parse(row.data) : null);
    });
  });
};

export const getAllScales = (): Promise<ScaleType[]> => {
  return new Promise((resolve, reject) => {
    db.all('SELECT data FROM scales', (err, rows: any[]) => {
      if (err) reject(err);
      else resolve(rows.map(row => JSON.parse(row.data)));
    });
  });
};
