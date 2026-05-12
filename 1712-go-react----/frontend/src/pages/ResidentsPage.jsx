import React, { useState, useEffect } from 'react';
import { residents } from '../api';
import { formatDate, calculateAge } from '../utils';

const ResidentsPage = ({ onSelectResident }) => {
  const [searchParams, setSearchParams] = useState({ name: '', idCard: '', community: '' });
  const [results, setResults] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [newResident, setNewResident] = useState({
    name: '',
    gender: 'male',
    idCard: '',
    birthDate: '',
    phone: '',
    address: '',
    emergencyContact: '',
    doctorInCharge: '',
    bloodType: 'A',
    allergies: [],
    community: '',
    guardian: { name: '', phone: '' },
  });
  const [newAllergy, setNewAllergy] = useState({ allergen: '', reaction: '' });

  const handleSearch = async () => {
    setLoading(true);
    setError('');
    try {
      const data = await residents.search(searchParams.name, searchParams.idCard, searchParams.community);
      setResults(data);
    } catch (err) {
      setError('搜索失败：' + err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    handleSearch();
  }, []);

  const handleCreateResident = async (e) => {
    e.preventDefault();
    setError('');
    try {
      const created = await residents.create(newResident);
      setResults([created, ...results]);
      setShowForm(false);
      setNewResident({
        name: '',
        gender: 'male',
        idCard: '',
        birthDate: '',
        phone: '',
        address: '',
        emergencyContact: '',
        doctorInCharge: '',
        bloodType: 'A',
        allergies: [],
        community: '',
        guardian: { name: '', phone: '' },
      });
    } catch (err) {
      setError('创建档案失败：' + err.message);
    }
  };

  const addAllergy = () => {
    if (newAllergy.allergen) {
      setNewResident({
        ...newResident,
        allergies: [...newResident.allergies, newAllergy],
      });
      setNewAllergy({ allergen: '', reaction: '' });
    }
  };

  return (
    <div className="page residents-page">
      <h1>居民档案管理</h1>

      <div className="search-panel">
        <h3>搜索条件</h3>
        <div className="search-form">
          <input
            type="text"
            placeholder="姓名"
            value={searchParams.name}
            onChange={(e) => setSearchParams({ ...searchParams, name: e.target.value })}
          />
          <input
            type="text"
            placeholder="身份证号"
            value={searchParams.idCard}
            onChange={(e) => setSearchParams({ ...searchParams, idCard: e.target.value })}
          />
          <input
            type="text"
            placeholder="社区"
            value={searchParams.community}
            onChange={(e) => setSearchParams({ ...searchParams, community: e.target.value })}
          />
          <button onClick={handleSearch} disabled={loading}>
            {loading ? '搜索中...' : '搜索'}
          </button>
          <button onClick={() => setShowForm(!showForm)} className="btn-secondary">
            {showForm ? '取消建档' : '新建档案'}
          </button>
        </div>
      </div>

      {error && <div className="error-message">{error}</div>}

      {showForm && (
        <div className="create-form">
          <h3>新建居民档案</h3>
          <form onSubmit={handleCreateResident}>
            <div className="form-row">
              <label>
                姓名：
                <input
                  type="text"
                  required
                  value={newResident.name}
                  onChange={(e) => setNewResident({ ...newResident, name: e.target.value })}
                />
              </label>
              <label>
                身份证号：
                <input
                  type="text"
                  required
                  value={newResident.idCard}
                  onChange={(e) => setNewResident({ ...newResident, idCard: e.target.value })}
                />
              </label>
            </div>
            <div className="form-row">
              <label>
                联系电话：
                <input
                  type="tel"
                  value={newResident.phone}
                  onChange={(e) => setNewResident({ ...newResident, phone: e.target.value })}
                />
              </label>
              <label>
                社区：
                <input
                  type="text"
                  value={newResident.community}
                  onChange={(e) => setNewResident({ ...newResident, community: e.target.value })}
                />
              </label>
            </div>
            <div className="form-row">
              <label>
                家庭住址：
                <input
                  type="text"
                  value={newResident.address}
                  onChange={(e) => setNewResident({ ...newResident, address: e.target.value })}
                />
              </label>
              <label>
                紧急联系人：
                <input
                  type="text"
                  value={newResident.emergencyContact}
                  onChange={(e) => setNewResident({ ...newResident, emergencyContact: e.target.value })}
                />
              </label>
            </div>
            <div className="form-row">
              <label>
                责任医生：
                <input
                  type="text"
                  value={newResident.doctorInCharge}
                  onChange={(e) => setNewResident({ ...newResident, doctorInCharge: e.target.value })}
                />
              </label>
              <label>
                血型：
                <select
                  value={newResident.bloodType}
                  onChange={(e) => setNewResident({ ...newResident, bloodType: e.target.value })}
                >
                  <option value="A">A</option>
                  <option value="B">B</option>
                  <option value="AB">AB</option>
                  <option value="O">O</option>
                </select>
              </label>
            </div>

            <div className="form-section">
              <h4>过敏史</h4>
              <div className="allergy-input">
                <input
                  type="text"
                  placeholder="过敏源"
                  value={newAllergy.allergen}
                  onChange={(e) => setNewAllergy({ ...newAllergy, allergen: e.target.value })}
                />
                <input
                  type="text"
                  placeholder="过敏反应"
                  value={newAllergy.reaction}
                  onChange={(e) => setNewAllergy({ ...newAllergy, reaction: e.target.value })}
                />
                <button type="button" onClick={addAllergy}>添加</button>
              </div>
              {newResident.allergies.length > 0 && (
                <ul className="allergy-list">
                  {newResident.allergies.map((a, i) => (
                    <li key={i}>{a.allergen} - {a.reaction}</li>
                  ))}
                </ul>
              )}
            </div>

            <div className="form-section">
              <h4>监护人信息（18岁以下必填）</h4>
              <div className="form-row">
                <label>
                  监护人姓名：
                  <input
                    type="text"
                    value={newResident.guardian.name}
                    onChange={(e) => setNewResident({
                      ...newResident,
                      guardian: { ...newResident.guardian, name: e.target.value },
                    })}
                  />
                </label>
                <label>
                  监护人电话：
                  <input
                    type="tel"
                    value={newResident.guardian.phone}
                    onChange={(e) => setNewResident({
                      ...newResident,
                      guardian: { ...newResident.guardian, phone: e.target.value },
                    })}
                  />
                </label>
              </div>
            </div>

            <button type="submit" className="btn-primary">创建档案</button>
          </form>
        </div>
      )}

      <div className="results-panel">
        <h3>搜索结果 ({results.length} 条)</h3>
        {results.length === 0 ? (
          <div className="empty-state">暂无居民档案</div>
        ) : (
          <table className="results-table">
            <thead>
              <tr>
                <th>姓名</th>
                <th>性别</th>
                <th>身份证号</th>
                <th>年龄</th>
                <th>社区</th>
                <th>责任医生</th>
                <th>建档日期</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {results.map((resident) => (
                <tr key={resident.id}>
                  <td>{resident.name}</td>
                  <td>{resident.gender === 'male' ? '男' : '女'}</td>
                  <td>{resident.idCard}</td>
                  <td>{calculateAge(resident.birthDate)}岁</td>
                  <td>{resident.community || '-'}</td>
                  <td>{resident.doctorInCharge || '-'}</td>
                  <td>{formatDate(resident.createDate)}</td>
                  <td>
                    <button onClick={() => onSelectResident && onSelectResident(resident)}>
                      查看详情
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};

export default ResidentsPage;
