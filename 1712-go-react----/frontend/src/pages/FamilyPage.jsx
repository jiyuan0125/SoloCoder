import React, { useState, useEffect } from 'react';
import { families, residents } from '../api';
import { formatDate, calculateAge } from '../utils';

const FamilyPage = () => {
  const [familyList, setFamilyList] = useState([]);
  const [selectedFamily, setSelectedFamily] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [newFamily, setNewFamily] = useState({
    familyNumber: '',
    address: '',
    contracted: false,
    members: [],
  });
  const [availableResidents, setAvailableResidents] = useState([]);
  const [newMemberID, setNewMemberID] = useState('');
  const [newMemberRelation, setNewMemberRelation] = useState('');

  const loadFamilies = async () => {
    setLoading(true);
    setError('');
    try {
      const data = await families.list();
      setFamilyList(data);
    } catch (err) {
      setError('加载家庭列表失败：' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const loadResidents = async () => {
    try {
      const data = await residents.search('', '', '');
      setAvailableResidents(data);
    } catch (err) {
      setError('加载居民列表失败：' + err.message);
    }
  };

  const loadFamilyMembers = async (family) => {
    if (!family) return;
    try {
      const membersWithInfo = await Promise.all(
        family.members.map(async (member) => {
          const residentList = await residents.search('', '', '');
          const matchingResident = residentList.find((r) => r.id === member.residentID);
          return { ...member, resident: matchingResident };
        })
      );
      return { ...family, members: membersWithInfo };
    } catch {
      return family;
    }
  };

  useEffect(() => {
    loadFamilies();
    loadResidents();
  }, []);

  const handleSelectFamily = async (family) => {
    const familyWithMembers = await loadFamilyMembers(family);
    setSelectedFamily(familyWithMembers);
  };

  const addMember = () => {
    if (newMemberID && newMemberRelation) {
      const resident = availableResidents.find((r) => r.id === newMemberID);
      if (resident) {
        setNewFamily({
          ...newFamily,
          members: [
            ...newFamily.members,
            { residentID: newMemberID, relation: newMemberRelation, resident },
          ],
        });
        setNewMemberID('');
        setNewMemberRelation('');
      }
    }
  };

  const handleCreateFamily = async (e) => {
    e.preventDefault();
    setError('');
    try {
      const familyData = {
        ...newFamily,
        members: newFamily.members.map((m) => ({
          residentID: m.residentID,
          relation: m.relation,
        })),
      };
      await families.create(familyData);
      setShowForm(false);
      loadFamilies();
      setNewFamily({
        familyNumber: '',
        address: '',
        contracted: false,
        members: [],
      });
    } catch (err) {
      setError('创建家庭档案失败：' + err.message);
    }
  };

  return (
    <div className="page family-page">
      <h1>家庭档案管理</h1>

      <div className="actions-bar">
        <button onClick={() => setShowForm(!showForm)} className="btn-primary">
          {showForm ? '取消' : '新建家庭档案'}
        </button>
      </div>

      {error && <div className="error-message">{error}</div>}

      {showForm && (
        <div className="create-form">
          <h3>新建家庭档案</h3>
          <form onSubmit={handleCreateFamily}>
            <div className="form-row">
              <label>
                家庭编号：
                <input
                  type="text"
                  required
                  value={newFamily.familyNumber}
                  onChange={(e) => setNewFamily({ ...newFamily, familyNumber: e.target.value })}
                />
              </label>
              <label>
                家庭住址：
                <input
                  type="text"
                  value={newFamily.address}
                  onChange={(e) => setNewFamily({ ...newFamily, address: e.target.value })}
                />
              </label>
            </div>
            <div className="form-row">
              <label>
                家庭医生签约状态：
                <select
                  value={newFamily.contracted ? 'yes' : 'no'}
                  onChange={(e) => setNewFamily({ ...newFamily, contracted: e.target.value === 'yes' })}
                >
                  <option value="no">未签约</option>
                  <option value="yes">已签约</option>
                </select>
              </label>
            </div>

            <div className="form-section">
              <h4>家庭成员</h4>
              <div className="member-input">
                <select
                  value={newMemberID}
                  onChange={(e) => setNewMemberID(e.target.value)}
                >
                  <option value="">选择居民</option>
                  {availableResidents.map((r) => (
                    <option key={r.id} value={r.id}>
                      {r.name} ({r.idCard})
                    </option>
                  ))}
                </select>
                <input
                  type="text"
                  placeholder="与户主关系"
                  value={newMemberRelation}
                  onChange={(e) => setNewMemberRelation(e.target.value)}
                />
                <button type="button" onClick={addMember}>添加成员</button>
              </div>
              {newFamily.members.length > 0 && (
                <ul className="member-list">
                  {newFamily.members.map((m, i) => (
                    <li key={i}>
                      {m.resident?.name} - {m.relation}
                    </li>
                  ))}
                </ul>
              )}
            </div>

            <button type="submit" className="btn-primary">创建家庭档案</button>
          </form>
        </div>
      )}

      <div className="families-panel">
        <h3>家庭列表 ({familyList.length} 个家庭)</h3>
        {loading && familyList.length === 0 ? (
          <div className="loading">加载中...</div>
        ) : familyList.length === 0 ? (
          <div className="empty-state">暂无家庭档案</div>
        ) : (
          <div className="family-grid">
            {familyList.map((family) => (
              <div
                key={family.id}
                className={`family-card ${selectedFamily?.id === family.id ? 'selected' : ''}`}
                onClick={() => handleSelectFamily(family)}
              >
                <div className="family-header">
                  <span className="family-number">{family.familyNumber}</span>
                  {family.contracted ? (
                    <span className="badge contracted">已签约</span>
                  ) : (
                    <span className="badge not-contracted">未签约</span>
                  )}
                </div>
                <div className="family-body">
                  <p><strong>地址：</strong>{family.address || '-'}</p>
                  <p><strong>成员数：</strong>{family.members.length} 人</p>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {selectedFamily && (
        <div className="family-detail">
          <h3>家庭详情</h3>
          <div className="detail-info">
            <p><strong>家庭编号：</strong>{selectedFamily.familyNumber}</p>
            <p><strong>地址：</strong>{selectedFamily.address || '-'}</p>
            <p><strong>签约状态：</strong>
              {selectedFamily.contracted ? (
                <span className="badge contracted">已签约</span>
              ) : (
                <span className="badge not-contracted">未签约</span>
              )}
            </p>
          </div>

          <h4>家庭成员 ({selectedFamily.members.length} 人)</h4>
          {selectedFamily.members.length === 0 ? (
            <div className="empty-state">暂无家庭成员</div>
          ) : (
            <div className="members-table">
              <table className="results-table">
                <thead>
                  <tr>
                    <th>姓名</th>
                    <th>关系</th>
                    <th>年龄</th>
                    <th>身份证号</th>
                    <th>责任医生</th>
                    <th>健康概况</th>
                  </tr>
                </thead>
                <tbody>
                  {selectedFamily.members.map((member) => (
                    <tr key={member.residentID}>
                      <td>{member.resident?.name || '未知'}</td>
                      <td>{member.relation}</td>
                      <td>
                        {member.resident ? `${calculateAge(member.resident.birthDate)}岁` : '-'}
                      </td>
                      <td>{member.resident?.idCard || '-'}</td>
                      <td>{member.resident?.doctorInCharge || '-'}</td>
                      <td>
                        {member.resident?.allergies?.length > 0 ? (
                          <span className="badge warning">有过敏史</span>
                        ) : (
                          <span className="badge normal">无特殊情况</span>
                        )}
                        {member.resident?.bloodType && (
                          <span className="badge info">
                            {member.resident.bloodType}型血
                          </span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default FamilyPage;
