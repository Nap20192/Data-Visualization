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



point_cloud = o3d.io.read_point_cloud("model.ply")
o3d.visualization.draw_geometries([point_cloud], window_name="Облако точек из модели")

print("=== Шаг 2: Облако точек из модели ===")
print("Количество точек:", np.asarray(point_cloud.points).shape[0])
print("Наличие цвета:", point_cloud.has_colors())

mesh_from_pc = o3d.geometry.TriangleMesh.create_from_point_cloud_poisson(point_cloud, depth=9)[0]
mesh_from_pc.crop(point_cloud.get_axis_aligned_bounding_box())
o3d.visualization.draw_geometries([mesh_from_pc], window_name="Модель из облака точек")

print("=== Шаг 3: Модель из облака точек ===")
print("Количество вершин:", np.asarray(mesh_from_pc.vertices).shape[0])
print("Количество треугольников:", np.asarray(mesh_from_pc.triangles).shape[0])
print("Наличие цвета:", mesh_from_pc.has_vertex_colors())

voxel_grid = o3d.geometry.VoxelGrid.create_from_point_cloud(point_cloud, voxel_size=0.05)
o3d.visualization.draw_geometries([voxel_grid], window_name="Воксельная сетка из облака точек")
print("=== Шаг 4: Воксельная сетка из облака точек ===")
print("Количество вокселей:", len(voxel_grid.get_voxels()))
print("Наличие цвета:", voxel_grid.has_colors())

bbox = mesh.get_axis_aligned_bounding_box()
min_bound = bbox.min_bound
max_bound = bbox.max_bound
center = bbox.get_center()
size = max_bound - min_bound

plane_margin = 0.2
plane_width = size[0] + plane_margin
plane_depth = size[2] + plane_margin
plane_height = 0.01

plane = o3d.geometry.TriangleMesh.create_box(width=plane_width,
                                             height=plane_height,
                                             depth=plane_depth)
plane.paint_uniform_color([0.8, 0.1, 0.1])

plane.translate([center[0] - plane_width/2,
                 min_bound[1] +25,
                 center[2] - plane_depth/2])

# Визуализируем модель с плоскостью
o3d.visualization.draw_geometries([mesh, plane], window_name="Модель с пересекающей плоскостью")


print("=== Шаг 5: Модель с пересекающей плоскостью ===")
# Используем центр плоскости для клиппинга
plane_center = np.asarray(plane.get_center())
plane_normal = np.array([0, 1, 0])
plane_point = plane_center

points = np.asarray(mesh.vertices)

distances = np.dot(points - plane_point, plane_normal)

mask = distances < 0
clipped_points = points[mask]

clipped_pcd = o3d.geometry.PointCloud()

clipped_pcd.points = o3d.utility.Vector3dVector(clipped_points)

if mesh.has_vertex_colors():
    colors = np.asarray(mesh.vertex_colors)
    clipped_pcd.colors = o3d.utility.Vector3dVector(colors[mask])

if not clipped_pcd.has_normals():
    clipped_pcd.estimate_normals()

# Визуализация
o3d.visualization.draw_geometries([clipped_pcd, plane], window_name="После обрезки по плоскости")

print("=== Шаг 6: Обрезка по плоскости ===")
print("Количество оставшихся точек:", np.asarray(clipped_pcd.points).shape[0])
print("Колличество оставшихся треугольников", np.asarray(clipped_pcd.triangles).shape[0])
print("Наличие цвета:", clipped_pcd.has_colors())
print("Наличие нормалей:", clipped_pcd.has_normals())

points = np.asarray(clipped_pcd.points)

# Убираем исходные цвета
clipped_pcd.colors = o3d.utility.Vector3dVector(np.zeros_like(points))

# Создаём градиент по Z
z_min, z_max = points[:,2].min(), points[:,2].max()
colors = np.zeros_like(points)
colors[:,2] = (points[:,2] - z_min) / (z_max - z_min)
clipped_pcd.colors = o3d.utility.Vector3dVector(colors)

# Находим экстремальные точки
z_min_idx = np.argmin(points[:,2])
z_max_idx = np.argmax(points[:,2])
extremes = [points[z_min_idx], points[z_max_idx]]

spheres = []
for pt in extremes:
    sphere = o3d.geometry.TriangleMesh.create_sphere(radius=2)
    sphere.translate(pt)
    sphere.paint_uniform_color([1,0,0])
    spheres.append(sphere)

o3d.visualization.draw_geometries([clipped_pcd] + spheres, window_name="Градиент и экстремумы")

print("=== Шаг 7: Градиент и экстремумы ===")
print("Минимум Z:", extremes[0])
print("Максимум Z:", extremes[1])
