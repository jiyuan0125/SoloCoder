import React, { useState, useEffect } from 'react'
import { drugApi, categoryApi } from '../api'

const dosageForms = [
  { value: 'tablet', label: '片剂' },
  { value: 'capsule', label: '胶囊' },
  { value: 'injection', label: '注射液' },
  { value: 'granule', label: '颗粒剂' },
  { value: 'syrup', label: '糖浆' },
  { value: 'external', label: '外用' },
]

const defaultDrug = {
  drug_code: '',
  generic_name: '',
  brand_name: '',
  specification: '',
  manufacturer: '',
  dosage_form: 'tablet',
  unit: '盒',
  retail_price: '',
  purchase_price: '',
  category_id: '',
  support_split: false,
  is_special_drug: false,
  min_stock: 10,
  max_stock: 1000,
}

export default function DrugCatalog() {
  const [drugs, setDrugs] = useState([])
  const [categories, setCategories] = useState([])
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingDrug, setEditingDrug] = useState(null)
  const [formData, setFormData] = useState(defaultDrug)
  const [error, setError] = useState('')

  const fetchDrugs = async () => {
    setLoading(true)
    try {
      const res = await drugApi.list(search)
      setDrugs(res.data.data || [])
    } catch (err) {
      console.error('获取药品列表失败:', err)
    } finally {
      setLoading(false)
    }
  }

  const fetchCategories = async () => {
    try {
      const res = await categoryApi.list()
      setCategories(res.data.data || [])
    } catch (err) {
      console.error('获取类别失败:', err)
    }
  }

  useEffect(() => {
    fetchDrugs()
    fetchCategories()
  }, [])

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchDrugs()
    }, 300)
    return () => clearTimeout(timer)
  }, [search])

  const handleOpenModal = (drug = null) => {
    setEditingDrug(drug)
    if (drug) {
      setFormData({ ...drug, category_id: drug.category_id || '' })
    } else {
      setFormData({ ...defaultDrug })
    }
    setError('')
    setModalOpen(true)
  }

  const handleCloseModal = () => {
    setModalOpen(false)
    setEditingDrug(null)
  }

  const handleInputChange = (e) => {
    const { name, value, type, checked } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }))
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    
    try {
      const data = {
        ...formData,
        retail_price: parseFloat(formData.retail_price) || 0,
        purchase_price: parseFloat(formData.purchase_price) || 0,
        min_stock: parseInt(formData.min_stock) || 0,
        max_stock: parseInt(formData.max_stock) || 0,
        category_id: formData.category_id ? parseInt(formData.category_id) : null,
      }

      if (editingDrug) {
        await drugApi.update(editingDrug.id, data)
      } else {
        await drugApi.create(data)
      }
      handleCloseModal()
      fetchDrugs()
    } catch (err) {
      setError(err.response?.data?.error || '操作失败')
    }
  }

  const handleDelete = async (drug) => {
    if (!confirm(`确定要删除药品【${drug.generic_name}】吗？`)) return
    try {
      await drugApi.remove(drug.id)
      fetchDrugs()
    } catch (err) {
      alert(err.response?.data?.error || '删除失败')
    }
  }

  const getDosageFormLabel = (value) => {
    const form = dosageForms.find(f => f.value === value)
    return form ? form.label : value
  }

  return (
    <div className="page-container">
      <h2 className="page-title">药品目录</h2>
      
      <div className="toolbar">
        <div className="search-box">
          <input
            type="text"
            className="search-input"
            placeholder="搜索药品编码、通用名、商品名..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <button className="btn btn-primary" onClick={() => handleOpenModal()}>
          + 新增药品
        </button>
      </div>

      <table className="data-table">
        <thead>
          <tr>
            <th>药品编码</th>
            <th>通用名</th>
            <th>商品名</th>
            <th>规格</th>
            <th>剂型</th>
            <th>单位</th>
            <th>零售价</th>
            <th>进货价</th>
            <th>当前库存</th>
            <th>标签</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          {loading ? (
            <tr><td colSpan="11" className="empty-state">加载中...</td></tr>
          ) : drugs.length === 0 ? (
            <tr><td colSpan="11" className="empty-state">暂无数据</td></tr>
          ) : (
            drugs.map(drug => (
              <tr key={drug.id}>
                <td>{drug.drug_code}</td>
                <td>{drug.generic_name}</td>
                <td>{drug.brand_name}</td>
                <td>{drug.specification}</td>
                <td>{getDosageFormLabel(drug.dosage_form)}</td>
                <td>{drug.unit}</td>
                <td>¥{drug.retail_price.toFixed(2)}</td>
                <td>¥{drug.purchase_price.toFixed(2)}</td>
                <td>{drug.current_stock}</td>
                <td>
                  {drug.is_special_drug && <span className="tag tag-critical">特殊药品</span>}
                  {drug.support_split && <span className="tag tag-near">可拆零</span>}
                </td>
                <td>
                  <button className="btn btn-primary btn-small" onClick={() => handleOpenModal(drug)}>
                    编辑
                  </button>
                  <button className="btn btn-danger btn-small" onClick={() => handleDelete(drug)}>
                    删除
                  </button>
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>

      {modalOpen && (
        <div className="modal-overlay" onClick={handleCloseModal}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>{editingDrug ? '编辑药品' : '新增药品'}</h3>
              <button className="modal-close" onClick={handleCloseModal}>&times;</button>
            </div>
            <form onSubmit={handleSubmit}>
              <div className="modal-body">
                {error && <div className="error-message">{error}</div>}
                
                <div className="form-row">
                  <div className="form-group">
                    <label className="form-label">药品编码 *</label>
                    <input
                      type="text"
                      className="form-input"
                      name="drug_code"
                      value={formData.drug_code}
                      onChange={handleInputChange}
                      placeholder="国药准字H11020101"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="form-label">通用名 *</label>
                    <input
                      type="text"
                      className="form-input"
                      name="generic_name"
                      value={formData.generic_name}
                      onChange={handleInputChange}
                      required
                    />
                  </div>
                </div>

                <div className="form-row">
                  <div className="form-group">
                    <label className="form-label">商品名</label>
                    <input
                      type="text"
                      className="form-input"
                      name="brand_name"
                      value={formData.brand_name}
                      onChange={handleInputChange}
                    />
                  </div>
                  <div className="form-group">
                    <label className="form-label">规格 *</label>
                    <input
                      type="text"
                      className="form-input"
                      name="specification"
                      value={formData.specification}
                      onChange={handleInputChange}
                      placeholder="如：10mg×24片"
                      required
                    />
                  </div>
                </div>

                <div className="form-row">
                  <div className="form-group">
                    <label className="form-label">生产厂家 *</label>
                    <input
                      type="text"
                      className="form-input"
                      name="manufacturer"
                      value={formData.manufacturer}
                      onChange={handleInputChange}
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="form-label">剂型 *</label>
                    <select
                      className="form-select"
                      name="dosage_form"
                      value={formData.dosage_form}
                      onChange={handleInputChange}
                    >
                      {dosageForms.map(f => (
                        <option key={f.value} value={f.value}>{f.label}</option>
                      ))}
                    </select>
                  </div>
                </div>

                <div className="form-row">
                  <div className="form-group">
                    <label className="form-label">单位 *</label>
                    <input
                      type="text"
                      className="form-input"
                      name="unit"
                      value={formData.unit}
                      onChange={handleInputChange}
                      placeholder="盒、瓶、支等"
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="form-label">类别</label>
                    <select
                      className="form-select"
                      name="category_id"
                      value={formData.category_id}
                      onChange={handleInputChange}
                    >
                      <option value="">请选择</option>
                      {categories.map(c => (
                        <option key={c.id} value={c.id}>{c.name}</option>
                      ))}
                    </select>
                  </div>
                </div>

                <div className="form-row">
                  <div className="form-group">
                    <label className="form-label">零售价 (元) *</label>
                    <input
                      type="number"
                      step="0.01"
                      className="form-input"
                      name="retail_price"
                      value={formData.retail_price}
                      onChange={handleInputChange}
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="form-label">进货价 (元) *</label>
                    <input
                      type="number"
                      step="0.01"
                      className="form-input"
                      name="purchase_price"
                      value={formData.purchase_price}
                      onChange={handleInputChange}
                      required
                    />
                  </div>
                </div>

                <div className="form-row">
                  <div className="form-group">
                    <label className="form-label">库存预警下限</label>
                    <input
                      type="number"
                      className="form-input"
                      name="min_stock"
                      value={formData.min_stock}
                      onChange={handleInputChange}
                    />
                  </div>
                  <div className="form-group">
                    <label className="form-label">库存预警上限</label>
                    <input
                      type="number"
                      className="form-input"
                      name="max_stock"
                      value={formData.max_stock}
                      onChange={handleInputChange}
                    />
                  </div>
                </div>

                <div className="form-row">
                  <div className="form-group">
                    <label style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                      <input
                        type="checkbox"
                        name="support_split"
                        checked={formData.support_split}
                        onChange={handleInputChange}
                      />
                      支持拆零销售
                    </label>
                  </div>
                  <div className="form-group">
                    <label style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                      <input
                        type="checkbox"
                        name="is_special_drug"
                        checked={formData.is_special_drug}
                        onChange={handleInputChange}
                      />
                      特殊药品（需双人复核）
                    </label>
                  </div>
                </div>
              </div>
              <div className="modal-footer">
                <button type="button" className="btn btn-default" onClick={handleCloseModal}>
                  取消
                </button>
                <button type="submit" className="btn btn-primary">
                  {editingDrug ? '保存' : '创建'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
