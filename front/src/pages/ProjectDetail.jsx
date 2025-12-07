import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { projectsAPI, teachersAPI } from '../api'

function ProjectDetail() {
    const { id } = useParams()
    const navigate = useNavigate()
    const [activeProject, setActiveProject] = useState(null)
    const [activeProjectRecords, setActiveProjectRecords] = useState([])
    const [teachers, setTeachers] = useState([])
    const [showAddMemberModal, setShowAddMemberModal] = useState(false)
    const [addMemberForm, setAddMemberForm] = useState({
        teacher_ids: [],
    })

    useEffect(() => {
        loadProject()
        loadTeachers()
    }, [id])

    const loadProject = async () => {
        try {
            const res = await projectsAPI.getById(id)
            setActiveProject(res.data?.data)
            const trackingRes = await projectsAPI.getTracking(id)
            setActiveProjectRecords(trackingRes.data?.data?.details || [])
        } catch (err) {
            console.error('加载项目失败：', err)
        }
    }

    const loadTeachers = async () => {
        try {
            const res = await teachersAPI.getAll()
            setTeachers(res.data?.data || [])
        } catch (err) {
            console.error('加载教师失败：', err)
        }
    }

    const dispatchEmails = async () => {
        if (!activeProject) return
        try {
            await projectsAPI.dispatch(activeProject.id)
            alert('邮件发送任务已启动！')
            loadProject()
        } catch (err) {
            alert('发送失败：' + (err.response?.data?.error || err.message))
        }
    }

    const refreshStatus = async () => {
        if (!activeProject) return
        try {
            await projectsAPI.fetchEmails(activeProject.id)
            const res = await projectsAPI.getTracking(activeProject.id)
            setActiveProjectRecords(res.data?.data?.details || [])
            alert('邮件已同步，状态已刷新！')
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

    const addMembers = async () => {
        if (!activeProject) return
        try {
            await projectsAPI.addMembers(activeProject.id, { teacher_ids: addMemberForm.teacher_ids })
            setShowAddMemberModal(false)
            setAddMemberForm({ teacher_ids: [] })
            refreshStatus()
            alert('成员添加成功！')
        } catch (err) {
            alert('添加失败：' + (err.response?.data?.error || err.message))
        }
    }

    if (!activeProject) return <div>加载中...</div>

    return (
        <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center mb-6 border-b pb-4">
                <button
                    onClick={() => navigate('/projects')}
                    className="mr-4 text-gray-500 hover:text-gray-700"
                >
                    <i className="fas fa-arrow-left"></i> 返回
                </button>
                <h2 className="text-2xl font-bold flex-1">{activeProject.name} - 管理面板</h2>
                <div className="space-x-2">
                    <button
                        onClick={() => setShowAddMemberModal(true)}
                        className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 text-sm"
                    >
                        <i className="fas fa-user-plus mr-1"></i> 添加成员
                    </button>
                    <button
                        onClick={() => dispatchEmails()}
                        className="bg-indigo-600 text-white px-4 py-2 rounded hover:bg-indigo-700 text-sm"
                    >
                        <i className="fas fa-paper-plane mr-1"></i> 发送邮件
                    </button>
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
                                        className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${record.status === 'replied'
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

            {/* Add Member Modal */}
            {showAddMemberModal && (
                <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full flex items-center justify-center z-50">
                    <div className="bg-white p-8 rounded-lg shadow-xl w-1/3">
                        <h3 className="text-xl font-bold mb-4">添加成员</h3>
                        <div className="space-y-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700">选择教师 (按住Ctrl多选)</label>
                                <select
                                    multiple
                                    className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm p-2 h-64"
                                    value={addMemberForm.teacher_ids}
                                    onChange={(e) => {
                                        const options = [...e.target.selectedOptions];
                                        const values = options.map(option => parseInt(option.value));
                                        setAddMemberForm({ ...addMemberForm, teacher_ids: values });
                                    }}
                                >
                                    {teachers.map((teacher) => (
                                        <option key={teacher.id} value={teacher.id}>
                                            {teacher.name} ({teacher.department_name || '未分配'})
                                        </option>
                                    ))}
                                </select>
                            </div>
                        </div>
                        <div className="mt-6 flex justify-end space-x-3">
                            <button
                                onClick={() => setShowAddMemberModal(false)}
                                className="bg-gray-200 text-gray-700 px-4 py-2 rounded hover:bg-gray-300"
                            >
                                取消
                            </button>
                            <button
                                onClick={addMembers}
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

export default ProjectDetail
