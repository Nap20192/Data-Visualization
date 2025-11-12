import open3d as o3d
import numpy as np

mesh = o3d.io.read_triangle_mesh("model.ply")
mesh.compute_vertex_normals()

print("=== Шаг 1: Исходная модель ===")
print("Количество вершин:", np.asarray(mesh.vertices).shape[0])
print("Количество треугольников:", np.asarray(mesh.triangles).shape[0])
print("Наличие цвета:", mesh.has_vertex_colors())
print("Наличие нормалей:", mesh.has_vertex_normals())

o3d.visualization.draw_geometries([mesh], window_name="Исходная модель")

pcd = mesh.sample_points_uniformly(number_of_points=5000)

print("\n=== Шаг 2: Облако точек ===")
print("Количество вершин (точек):", np.asarray(pcd.points).shape[0])
print("Наличие цвета:", pcd.has_colors())

o3d.visualization.draw_geometries([pcd], window_name="Облако точек")

mesh_poisson, _ = o3d.geometry.TriangleMesh.create_from_point_cloud_poisson(pcd, depth=8)
bbox = mesh.get_axis_aligned_bounding_box()
mesh_poisson = mesh_poisson.crop(bbox)

print("\n=== Шаг 3: Реконструированный mesh ===")
print("Количество вершин:", np.asarray(mesh_poisson.vertices).shape[0])
print("Количество треугольников:", np.asarray(mesh_poisson.triangles).shape[0])
print("Наличие цвета:", mesh_poisson.has_vertex_colors())

o3d.visualization.draw_geometries([mesh_poisson], window_name="Реконструкция Poisson")

voxel_size = 0.1
voxel_grid = o3d.geometry.VoxelGrid.create_from_point_cloud(pcd, voxel_size=voxel_size)

print("\n=== Шаг 4: Вокселизация ===")
print("Количество вершин (центров вокселей):", len(voxel_grid.get_voxels()))
print("Наличие цвета:", voxel_grid.has_colors())

o3d.visualization.draw_geometries([voxel_grid], window_name="Воксели")

plane = o3d.geometry.TriangleMesh.create_box(width=1.0, height=0.01, depth=1.0)
plane.paint_uniform_color([0.8, 0.1, 0.1])
plane.translate(mesh.get_center() + np.array([0, -0.5, 0]))

print("\n=== Шаг 5: Плоскость ===")
o3d.visualization.draw_geometries([mesh, plane], window_name="Объект + плоскость")

# Простая обрезка по оси Y
points = np.asarray(pcd.points)
mask = points[:,1] > plane.get_center()[1]  # оставляем точки выше плоскости
pcd_clipped = o3d.geometry.PointCloud()
pcd_clipped.points = o3d.utility.Vector3dVector(points[mask])

print("\n=== Шаг 6: Обрезка ===")
print("Количество оставшихся точек:", np.asarray(pcd_clipped.points).shape[0])
print("Наличие цвета:", pcd_clipped.has_colors())
print("Наличие нормалей:", pcd_clipped.has_normals())

o3d.visualization.draw_geometries([pcd_clipped, plane], window_name="Обрезка по плоскости")

# Убираем цвета
pcd_clipped.colors = o3d.utility.Vector3dVector(np.zeros_like(np.asarray(pcd_clipped.points)))

# Задаём градиент по Z
points = np.asarray(pcd_clipped.points)
colors = np.zeros_like(points)
colors[:, 2] = (points[:, 2] - points[:, 2].min()) / (points[:, 2].ptp())
pcd_clipped.colors = o3d.utility.Vector3dVector(colors)

# Экстремальные точки по Z
z_min_idx = np.argmin(points[:,2])
z_max_idx = np.argmax(points[:,2])
extremes = [points[z_min_idx], points[z_max_idx]]

spheres = []
for pt in extremes:
    s = o3d.geometry.TriangleMesh.create_sphere(radius=0.05)
    s.translate(pt)
    s.paint_uniform_color([1,0,0])
    spheres.append(s)

print("\n=== Шаг 7: Цвет и экстремумы ===")
print("Минимум Z:", extremes[0])
print("Максимум Z:", extremes[1])

o3d.visualization.draw_geometries([pcd_clipped] + spheres, window_name="Градиент и экстремумы")
