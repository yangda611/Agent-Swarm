import {useEffect, useRef} from "react";
import * as echarts from "echarts/core";
import {BarChart, GraphChart, PieChart, ScatterChart} from "echarts/charts";
import {GridComponent, LegendComponent, TooltipComponent} from "echarts/components";
import {SVGRenderer} from "echarts/renderers";
import type {EChartsOption} from "echarts";

echarts.use([GridComponent, LegendComponent, TooltipComponent, PieChart, BarChart, ScatterChart, GraphChart, SVGRenderer]);

interface EChartPanelProps {
    className?: string;
    minHeight?: number;
    option: EChartsOption;
}

export function EChartPanel({className = "", minHeight = 320, option}: EChartPanelProps) {
    const chartRef = useRef<HTMLDivElement | null>(null);

    useEffect(() => {
        if (!chartRef.current) {
            return undefined;
        }

        const chart = echarts.getInstanceByDom(chartRef.current)
            || echarts.init(chartRef.current, undefined, {renderer: "svg"});
        chart.setOption(option, true);

        const resizeObserver = new ResizeObserver(() => {
            chart.resize();
        });
        resizeObserver.observe(chartRef.current);

        return () => {
            resizeObserver.disconnect();
            chart.dispose();
        };
    }, [option]);

    return (
        <div
            className={`chart-surface ${className}`.trim()}
            ref={chartRef}
            style={{minHeight}}
        />
    );
}
