import { useState, useEffect } from 'react'
import { projectsAPI, teachersAPI, departmentsAPI } from './api'

function App() {
  const [currentView, setCurrentView] = useState('projects')
  const [projects, setProjects] = useState([])
  const [teachers, setTeachers] = useState([])
  const [departments, setDepartments] = useState([])
  const [activeProject, setActiveProject] = useState(null)
  const [activeProjectRecords, setActiveProjectRecords] = useState([])
  const [showCreateProjectModal, setShowCreateProjectModal] = useState(false)
  const [showCreateTeacherModal, setShowCreateTeacherModal] = useState(false)
  
  // Form states
  const [projectForm, setProjectForm] = useState({
    name: '',
    code: '',
    email_subject_template: '',
    email_body_template: '',
    excel_template: null,
  })
  
  const [teacherForm, setTeacherForm] = useState({
    name: '',
    email: '',
    department_id: '',
    phone: '',
  })
  
  const [dispatchForm, setDispatchForm] = useState({
    target_type: 'all',
    department_id: '',
    teacher_ids: [],
  })

  useEffect(() => {
    loadProjects()
    loadTeachers()
    loadDepartments()
  }, [])

  const loadProjects = async () => {
    try {
      const res = await projectsAPI.getAll()
      setProjects(res.data?.data || [])
    } catch (err) {
      console.error('Failed to load projects:', err)
    }
  }

  const loadTeachers = async () => {
    try {
      const res = await teachersAPI.getAll()
      setTeachers(res.data?.data || [])
    } catch (err) {
      console.error('Failed to load teachers:', err)
    }
  }

  const loadDepartments = async () => {
    try {
      const res = await departmentsAPI.getAll()
      setDepartments(res.data?.data || [])
    } catch (err) {
      console.error('Failed to load departments:', err)
    }
  }

  const openProjectDetail = async (project) => {
    setActiveProject(project)
    setCurrentView('project_detail')
    try {
      const res = await projectsAPI.getTracking(project.id)
      setActiveProjectRecords(res.data?.data?.details || [])
    } catch (err) {
      console.error('Failed to load tracking:', err)
    }
  }

  const createProject = async () => {
    try {
      const formData = new FormData()
      formData.append('name', projectForm.name)
      formData.append('code', projectForm.code)
      formData.append('email_subject_template', projectForm.email_subject_template)
      formData.append('email_body_template', projectForm.email_body_template)
      if (projectForm.excel_template) {
        formData.append('excel_template', projectForm.excel_template)
      }
      
      await projectsAPI.create(formData)
      setShowCreateProjectModal(false)
      setProjectForm({ name: '', code: '', email_subject_template: '', email_body_template: '', excel_template: null })
      loadProjects()
      alert('项目创建成功！')
    } catch (err) {
      alert('创建失败：' + (err.response?.data?.error || err.message))
    }
  }

  const dispatchEmails = async () => {
    if (!activeProject) return
    try {
      await projectsAPI.dispatch(activeProject.id, dispatchForm)
      alert('邮件发送成功！')
      loadProjects()
    } catch (err) {
      alert('发送失败：' + (err.response?.data?.error || err.message))
    }
  }

  const refreshStatus = async () => {
    if (!activeProject) return
    try {
      const res = await projectsAPI.getTracking(activeProject.id)
      setActiveProjectRecords(res.data?.data?.details || [])
      alert('状态已刷新！')
    } catch (err) {
      alert('刷新失败：' + (err.response?.data?.error || err.message))
    }
  }

  const remindAll = async () => {
    if (!activeProject) return
    try {
      await projectsAPI.remind(activeProject.id, {})
      alert('催办邮件已发送！')
    } catch (err) {
      alert('催办失败：' + (err.response?.data?.error || err.message))
    }
  }

  const remindOne = async (record) => {
    if (!activeProject) return
    try {
      await projectsAPI.remind(activeProject.id, { target_ids: [record.teacher_id] })
      alert(`已向 ${record.name} 发送催办邮件！`)
    } catch (err) {
      alert('催办失败：' + (err.response?.data?.error || err.message))
    }
  }

  const aggregateData = async () => {
    if (!activeProject) return
    try {
      await projectsAPI.aggregate(activeProject.id)
      alert('数据汇总中，请稍后下载...')
      setTimeout(async () => {
        const res = await projectsAPI.download(activeProject.id)
        const url = window.URL.createObjectURL(new Blob([res.data]))
        const link = document.createElement('a')
        link.href = url
        link.setAttribute('download', `${activeProject.name}_汇总.xlsx`)
        document.body.appendChild(link)
        link.click()
        link.remove()
      }, 2000)
    } catch (err) {
      alert('汇总失败：' + (err.response?.data?.error || err.message))
    }
  }

  const createTeacher = async () => {
    try {
      const data = {
        ...teacherForm,
        department_id: teacherForm.department_id ? parseInt(teacherForm.department_id) : null
      }
      await teachersAPI.create(data)
      setShowCreateTeacherModal(false)
      setTeacherForm({ name: '', email: '', department_id: '', phone: '' })
      loadTeachers()
      alert('教师添加成功！')
    } catch (err) {
      alert('添加失败：' + (err.response?.data?.error || err.message))
    }
  }

  return (
    <div className="flex h-screen bg-gray-50">
      {/* Sidebar */}
      <aside className="w-64 bg-white shadow-md flex flex-col">
        <div className="p-6 border-b">
          <h1 className="text-xl font-bold text-blue-600">
            <i className="fas fa-chart-pie mr-2"></i>科研数据汇总
          </h1>
        </div>
        <nav className="flex-1 p-4 space-y-2">
          <button
            onClick={() => setCurrentView('projects')}
            className={`w-full text-left block px-4 py-2 rounded hover:bg-gray-100 transition ${
              currentView === 'projects' ? 'bg-blue-50 text-blue-600' : ''
            }`}
          >
            <i className="fas fa-folder-open mr-2"></i> 项目管理
          </button>
          <button
            onClick={() => setCurrentView('teachers')}
            className={`w-full text-left block px-4 py-2 rounded hover:bg-gray-100 transition ${
              currentView === 'teachers' ? 'bg-blue-50 text-blue-600' : ''
            }`}
          >
            <i className="fas fa-users mr-2"></i> 教师信息库
          </button>
        </nav>
        <div className="p-4 border-t text-xs text-gray-400">Version 1.0</div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-auto p-8">
        {/* Projects View */}
        {currentView === 'projects' && (
          <div>
            <div className="flex justify-between items-center mb-6">
              <h2 className="text-2xl font-bold">项目列表</h2>
              <button
                onClick={() => setShowCreateProjectModal(true)}
                className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 transition"
              >
                <i className="fas fa-plus mr-1"></i> 新建汇总项目
              </button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {projects.map((project) => (
                <div
                  key={project.id}
                  className="bg-white p-6 rounded-lg shadow hover:shadow-md transition border border-gray-100"
                >
                  <div className="flex justify-between items-start mb-4">
                    <h3 className="font-bold text-lg">{project.name}</h3>
                    <span
                      className={`px-2 py-1 rounded text-xs ${
                        project.status === 'active'
                          ? 'bg-green-100 text-green-800'
                          : 'bg-gray-100 text-gray-800'
                      }`}
                    >
                      {project.status === 'active' ? '进行中' : '已归档'}
                    </span>
                  </div>
                  <p className="text-sm text-gray-500 mb-4">
                    创建时间: {new Date(project.created_at).toLocaleDateString()}
                  </p>
                  <div className="flex justify-between items-center">
                    <div className="text-sm">
                      <span className="font-bold text-blue-600">{project.replied_count || 0}</span> /{' '}
                      {project.total_sent || 0} 已回复
                    </div>
                    <button
                      onClick={() => openProjectDetail(project)}
                      className="text-blue-600 hover:underline text-sm"
                    >
                      进入管理 &rarr;
                    </button>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2 mt-2">
                    <div
                      className="bg-blue-600 h-2 rounded-full"
                      style={{
                        width: `${
                          project.total_sent > 0
                            ? ((project.replied_count || 0) / project.total_sent) * 100
                            : 0
                        }%`,
                      }}
                    ></div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Project Detail View */}
        {currentView === 'project_detail' && activeProject && (
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center mb-6 border-b pb-4">
              <button
                onClick={() => setCurrentView('projects')}
                className="mr-4 text-gray-500 hover:text-gray-700"
              >
                <i className="fas fa-arrow-left"></i> 返回
              </button>
              <h2 className="text-2xl font-bold flex-1">{activeProject.name} - 管理面板</h2>
              <div className="space-x-2">
                <button
                  onClick={aggregateData}
                  className="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700 text-sm"
                >
                  <i className="fas fa-file-excel mr-1"></i> 汇总并下载数据
                </button>
              </div>
            </div>

            <div className="grid grid-cols-3 gap-6 mb-8">
              <div className="bg-blue-50 p-4 rounded border border-blue-100">
                <div className="text-gray-500 text-sm">总发送</div>
                <div className="text-2xl font-bold text-blue-700">{activeProject.total_sent || 0}</div>
              </div>
              <div className="bg-green-50 p-4 rounded border border-green-100">
                <div className="text-gray-500 text-sm">已回复</div>
                <div className="text-2xl font-bold text-green-700">{activeProject.replied_count || 0}</div>
              </div>
              <div className="bg-red-50 p-4 rounded border border-red-100">
                <div className="text-gray-500 text-sm">未回复</div>
                <div className="text-2xl font-bold text-red-700">
                  {(activeProject.total_sent || 0) - (activeProject.replied_count || 0)}
                </div>
              </div>
            </div>

            <div className="flex justify-between items-center mb-4">
              <h3 className="font-bold text-lg">回复状态监控</h3>
              <div className="space-x-2">
                <button
                  onClick={refreshStatus}
                  className="text-gray-600 hover:text-blue-600 px-3 py-1 border rounded text-sm"
                >
                  <i className="fas fa-sync-alt mr-1"></i> 刷新状态
                </button>
                <button
                  onClick={remindAll}
                  className="bg-orange-500 text-white px-3 py-1 rounded hover:bg-orange-600 text-sm"
                >
                  <i className="fas fa-bell mr-1"></i> 一键催办未回复
                </button>
              </div>
            </div>

            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      教师姓名
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      所在系
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      状态
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      回复时间
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      操作
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {activeProjectRecords.map((record) => (
                    <tr key={record.teacher_id}>
                      <td className="px-6 py-4 whitespace-nowrap">{record.name}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-gray-500">{record.department}</td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span
                          className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                            record.status === 'replied'
                              ? 'bg-green-100 text-green-800'
                              : 'bg-red-100 text-red-800'
                          }`}
                        >
                          {record.status === 'replied' ? '已回复' : '未回复'}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {record.reply_time || '-'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                        {record.status !== 'replied' ? (
                          <button
                            onClick={() => remindOne(record)}
                            className="text-orange-600 hover:text-orange-900"
                          >
                            催办
                          </button>
                        ) : (
                          <span className="text-gray-400">已完成</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* Teachers View */}
        {currentView === 'teachers' && (
          <div>
            <div className="flex justify-between items-center mb-6">
              <h2 className="text-2xl font-bold">教师信息库</h2>
              <button
                onClick={() => setShowCreateTeacherModal(true)}
                className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 transition"
              >
                <i className="fas fa-user-plus mr-1"></i> 添加教师
              </button>
            </div>
            <div className="bg-white shadow overflow-hidden border-b border-gray-200 sm:rounded-lg">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      姓名
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      系别
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      邮箱
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      电话
                    </th>
                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">
                      操作
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {teachers.map((teacher) => (
                    <tr key={teacher.id}>
                      <td className="px-6 py-4 whitespace-nowrap font-medium text-gray-900">
                        {teacher.name}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-gray-500">
                        {teacher.department_name || '-'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-gray-500">{teacher.email}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-gray-500">{teacher.phone || '-'}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        <button className="text-indigo-600 hover:text-indigo-900">编辑</button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </main>

      {/* Create Project Modal */}
      {showCreateProjectModal && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full flex items-center justify-center z-50">
          <div className="bg-white p-8 rounded-lg shadow-xl w-1/2 max-h-[90vh] overflow-y-auto">
            <h3 className="text-xl font-bold mb-4">新建汇总项目</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700">项目名称</label>
                <input
                  type="text"
                  value={projectForm.name}
                  onChange={(e) => setProjectForm({ ...projectForm, name: e.target.value })}
                  className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2"
                  placeholder="例如：2025年度科研工作量汇总"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">项目代码</label>
                <input
                  type="text"
                  value={projectForm.code}
                  onChange={(e) => setProjectForm({ ...projectForm, code: e.target.value })}
                  className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2"
                  placeholder="例如：2025_WORKLOAD"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">邮件标题模板</label>
                <input
                  type="text"
                  value={projectForm.email_subject_template}
                  onChange={(e) =>
                    setProjectForm({ ...projectForm, email_subject_template: e.target.value })
                  }
                  className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2"
                  placeholder="请回复：2025年度科研工作量汇总"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">邮件正文模板</label>
                <textarea
                  value={projectForm.email_body_template}
                  onChange={(e) =>
                    setProjectForm({ ...projectForm, email_body_template: e.target.value })
                  }
                  className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2"
                  rows="4"
                  placeholder="各位老师，请填写附件中的表格..."
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">上传Excel模板</label>
                <input
                  type="file"
                  accept=".xlsx,.xls"
                  onChange={(e) =>
                    setProjectForm({ ...projectForm, excel_template: e.target.files[0] })
                  }
                  className="mt-1 block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
                />
              </div>
            </div>
            <div className="mt-6 flex justify-end space-x-3">
              <button
                onClick={() => setShowCreateProjectModal(false)}
                className="bg-gray-200 text-gray-700 px-4 py-2 rounded hover:bg-gray-300"
              >
                取消
              </button>
              <button
                onClick={createProject}
                className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
              >
                创建项目
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Create Teacher Modal */}
      {showCreateTeacherModal && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full flex items-center justify-center z-50">
          <div className="bg-white p-8 rounded-lg shadow-xl w-1/3">
            <h3 className="text-xl font-bold mb-4">添加教师</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700">姓名</label>
                <input
                  type="text"
                  value={teacherForm.name}
                  onChange={(e) => setTeacherForm({ ...teacherForm, name: e.target.value })}
                  className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">邮箱</label>
                <input
                  type="email"
                  value={teacherForm.email}
                  onChange={(e) => setTeacherForm({ ...teacherForm, email: e.target.value })}
                  className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">所在系</label>
                <select
                  value={teacherForm.department_id}
                  onChange={(e) => setTeacherForm({ ...teacherForm, department_id: e.target.value })}
                  className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2"
                >
                  <option value="">选择系别</option>
                  {departments.map((dept) => (
                    <option key={dept.id} value={dept.id}>
                      {dept.name}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">电话</label>
                <input
                  type="text"
                  value={teacherForm.phone}
                  onChange={(e) => setTeacherForm({ ...teacherForm, phone: e.target.value })}
                  className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2"
                />
              </div>
            </div>
            <div className="mt-6 flex justify-end space-x-3">
              <button
                onClick={() => setShowCreateTeacherModal(false)}
                className="bg-gray-200 text-gray-700 px-4 py-2 rounded hover:bg-gray-300"
              >
                取消
              </button>
              <button
                onClick={createTeacher}
                className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
              >
                添加
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default App
