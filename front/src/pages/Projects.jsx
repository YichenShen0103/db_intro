import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { projectsAPI, userAPI } from '../api'

function Projects() {
    const navigate = useNavigate()
    const [projects, setProjects] = useState([])
    const [showCreateProjectModal, setShowCreateProjectModal] = useState(false)
    const [projectForm, setProjectForm] = useState({
        name: '',
        code: '',
        email_subject_template: '',
        email_body_template: '',
        excel_template: null,
    })

    useEffect(() => {
        checkEmailConfig()
        loadProjects()
    }, [])

    const checkEmailConfig = async () => {
        try {
            const res = await userAPI.getEmailConfig()
            if (!res.data.has_config) {
                alert("请先配置您的邮箱设置。")
                navigate('/settings')
            }
        } catch (err) {
            console.error("检查邮箱配置失败", err)
        }
    }

    const loadProjects = async () => {
        try {
            const res = await projectsAPI.getAll()
            setProjects(res.data?.data || [])
        } catch (err) {
            console.error('加载项目失败：', err)
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

    const dispatchEmails = async (project) => {
        if (!project) return
        try {
            await projectsAPI.dispatch(project.id)
            alert('邮件发送任务已启动！')
            loadProjects()
        } catch (err) {
            alert('发送失败：' + (err.response?.data?.error || err.message))
        }
    }

    return (
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
                                className={`px-2 py-1 rounded text-xs ${project.status === 'active'
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
                            <div className="space-x-2">
                                <button
                                    onClick={() => dispatchEmails(project)}
                                    className="text-indigo-600 hover:text-indigo-800 text-sm"
                                    title="发送邮件给新成员"
                                >
                                    <i className="fas fa-paper-plane mr-1"></i> 发送
                                </button>
                                <button
                                    onClick={() => navigate(`/projects/${project.id}`)}
                                    className="text-blue-600 hover:underline text-sm"
                                >
                                    进入管理 &rarr;
                                </button>
                            </div>
                        </div>
                        <div className="w-full bg-gray-200 rounded-full h-2 mt-2">
                            <div
                                className="bg-blue-600 h-2 rounded-full"
                                style={{
                                    width: `${project.total_sent > 0
                                        ? ((project.replied_count || 0) / project.total_sent) * 100
                                        : 0
                                        }%`,
                                }}
                            ></div>
                        </div>
                    </div>
                ))}
            </div>

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
        </div>
    )
}

export default Projects
