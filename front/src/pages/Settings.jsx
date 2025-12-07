import React, { useState, useEffect } from 'react';
import { userAPI } from '../api';
import { useNavigate } from 'react-router-dom';

const Settings = () => {
    const [config, setConfig] = useState({
        smtp_host: '',
        smtp_port: '',
        smtp_username: '',
        smtp_password: '',
        imap_host: '',
        imap_port: '',
        imap_username: '',
        imap_password: '',
        email_address: ''
    });
    const [loading, setLoading] = useState(true);
    const [message, setMessage] = useState('');
    const navigate = useNavigate();

    useEffect(() => {
        fetchConfig();
    }, []);

    const fetchConfig = async () => {
        try {
            const res = await userAPI.getEmailConfig();
            // If config exists, populate fields. Passwords are not returned, so keep them empty.
            setConfig(prev => ({
                ...prev,
                ...res.data,
                smtp_password: '',
                imap_password: ''
            }));
        } catch (error) {
            console.error("Failed to fetch config", error);
        } finally {
            setLoading(false);
        }
    };

    const handleChange = (e) => {
        const { name, value } = e.target;
        setConfig(prev => ({ ...prev, [name]: value }));
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            await userAPI.updateEmailConfig(config);
            setMessage('配置保存成功！');
            // Optionally redirect to dashboard if they came from there
        } catch (error) {
            setMessage('保存配置失败。');
        }
    };

    if (loading) return <div>加载中...</div>;

    return (
        <div>
            <h1 className="text-2xl font-bold mb-4">邮箱设置</h1>
            {message && <div className="mb-4 p-2 bg-blue-100 text-blue-700 rounded">{message}</div>}
            <form onSubmit={handleSubmit} className="bg-white shadow-md rounded px-8 pt-6 pb-8 mb-4">
                <div className="mb-4">
                    <label className="block text-gray-700 text-sm font-bold mb-2">邮箱地址 (发件人)</label>
                    <input className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                        name="email_address" value={config.email_address} onChange={handleChange} required />
                </div>

                <h2 className="text-xl font-bold mb-2 mt-4">SMTP 设置 (发送)</h2>
                <div className="grid grid-cols-2 gap-4">
                    <div className="mb-4">
                        <label className="block text-gray-700 text-sm font-bold mb-2">SMTP 服务器</label>
                        <input className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                            name="smtp_host" value={config.smtp_host} onChange={handleChange} required />
                    </div>
                    <div className="mb-4">
                        <label className="block text-gray-700 text-sm font-bold mb-2">SMTP 端口</label>
                        <input className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                            name="smtp_port" value={config.smtp_port} onChange={handleChange} required />
                    </div>
                    <div className="mb-4">
                        <label className="block text-gray-700 text-sm font-bold mb-2">SMTP 用户名</label>
                        <input className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                            name="smtp_username" value={config.smtp_username} onChange={handleChange} required />
                    </div>
                    <div className="mb-4">
                        <label className="block text-gray-700 text-sm font-bold mb-2">SMTP 密码</label>
                        <input type="password" className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                            name="smtp_password" value={config.smtp_password} onChange={handleChange} placeholder="留空以保持不变" />
                    </div>
                </div>

                <h2 className="text-xl font-bold mb-2 mt-4">IMAP 设置 (接收)</h2>
                <div className="grid grid-cols-2 gap-4">
                    <div className="mb-4">
                        <label className="block text-gray-700 text-sm font-bold mb-2">IMAP 服务器</label>
                        <input className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                            name="imap_host" value={config.imap_host} onChange={handleChange} required />
                    </div>
                    <div className="mb-4">
                        <label className="block text-gray-700 text-sm font-bold mb-2">IMAP 端口</label>
                        <input className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                            name="imap_port" value={config.imap_port} onChange={handleChange} required />
                    </div>
                    <div className="mb-4">
                        <label className="block text-gray-700 text-sm font-bold mb-2">IMAP 用户名</label>
                        <input className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                            name="imap_username" value={config.imap_username} onChange={handleChange} required />
                    </div>
                    <div className="mb-4">
                        <label className="block text-gray-700 text-sm font-bold mb-2">IMAP 密码</label>
                        <input type="password" className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                            name="imap_password" value={config.imap_password} onChange={handleChange} placeholder="留空以保持不变" />
                    </div>
                </div>

                <div className="flex items-center justify-between mt-6">
                    <button className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline" type="submit">
                        保存配置
                    </button>
                </div>
            </form>
        </div>
    );
};

export default Settings;
