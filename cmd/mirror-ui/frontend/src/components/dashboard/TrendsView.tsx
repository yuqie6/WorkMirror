import React from 'react';

const TrendsView: React.FC = () => {
    // 模拟数据
    const weekData = [
        { day: '周一', value: 65 },
        { day: '周二', value: 80 },
        { day: '周三', value: 45 },
        { day: '周四', value: 90 },
        { day: '周五', value: 70 },
        { day: '周六', value: 30 },
        { day: '周日', value: 50 },
    ];

    return (
        <div className="space-y-8 pb-12 animate-slide-up">
            {/* 头部 */}
            <header>
                <h1 className="text-4xl font-bold text-gray-900 tracking-tight">
                    趋势分析
                </h1>
                <p className="text-gray-500 mt-1">Trends & Analytics</p>
            </header>

            {/* 主图表区 */}
            <div className="grid grid-cols-12 gap-5">
                {/* 周活动图表 */}
                <div className="col-span-8">
                    <div className="card">
                        <div className="flex items-center justify-between mb-6">
                            <div>
                                <h3 className="text-lg font-semibold text-gray-900">本周编码活动</h3>
                                <p className="text-sm text-gray-400">Weekly Coding Activity</p>
                            </div>
                            <div className="flex items-center gap-2">
                                <button className="pill pill-active">周</button>
                                <button className="pill">月</button>
                                <button className="pill">年</button>
                            </div>
                        </div>
                        
                        {/* 简易柱状图 */}
                        <div className="flex items-end justify-between gap-4 h-48 px-4">
                            {weekData.map((item, i) => (
                                <div key={i} className="flex-1 flex flex-col items-center gap-2">
                                    <div 
                                        className="w-full bg-gradient-to-t from-accent-gold to-amber-300 rounded-t-lg transition-all duration-500 hover:from-accent-gold hover:to-amber-200"
                                        style={{ height: `${item.value}%` }}
                                    />
                                    <span className="text-xs text-gray-400">{item.day}</span>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>

                {/* 统计摘要 */}
                <div className="col-span-4 space-y-5">
                    <div className="card-dark">
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-sm text-gray-400">本周总时长</span>
                            <svg className="w-5 h-5 text-accent-gold" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                        </div>
                        <div className="text-3xl font-bold text-white">12.5h</div>
                        <div className="text-sm text-green-400 mt-1">+23% vs 上周</div>
                    </div>

                    <div className="card">
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-sm text-gray-400">日均编码</span>
                            <svg className="w-5 h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" />
                            </svg>
                        </div>
                        <div className="text-3xl font-bold text-gray-900">1.8h</div>
                    </div>

                    <div className="card">
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-sm text-gray-400">最活跃日</span>
                        </div>
                        <div className="text-xl font-bold text-gray-900">周四</div>
                        <div className="text-sm text-gray-400">3.2 小时</div>
                    </div>
                </div>

                {/* 提示卡片 */}
                <div className="col-span-12">
                    <div className="card bg-gradient-to-r from-amber-50 to-orange-50 border border-amber-100">
                        <div className="flex items-center gap-4">
                            <div className="w-12 h-12 rounded-2xl bg-gradient-gold flex items-center justify-center text-white">
                                <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                                    <path strokeLinecap="round" strokeLinejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09z" />
                                </svg>
                            </div>
                            <div>
                                <h3 className="font-semibold text-gray-900">数据可视化即将上线</h3>
                                <p className="text-sm text-gray-500">
                                    Mirror 正在收集更多成长数据。完整的趋势图表、技能热力图和成长轨迹即将推出。
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default TrendsView;
